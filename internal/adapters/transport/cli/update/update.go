package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fkteams/internal/runtime/env"

	"github.com/Masterminds/semver/v3"
	"github.com/wsshow/selfupdate"
)

const (
	maxReleaseMetadataBytes int64 = 4 << 20
	maxUpdateArchiveBytes   int64 = 512 << 20
)

// Release 表示软件版本发布信息。
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset 表示发布中的可下载资源。
type Asset struct {
	Name               string `json:"name"`
	ContentType        string `json:"content_type"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// IsCompressedFile 判断资源是否为压缩文件。
func (a Asset) IsCompressedFile() bool {
	if a.ContentType == "application/zip" || a.ContentType == "application/x-gzip" {
		return true
	}
	parsed, err := url.Parse(a.BrowserDownloadURL)
	return err == nil && strings.HasSuffix(strings.ToLower(parsed.Path), ".zip")
}

type Updater struct {
	name string // 可执行文件名称（不含扩展名）
}

// NewUpdater 创建 Updater 实例。
func NewUpdater(name string) *Updater {
	return &Updater{name: name}
}

// CheckForUpdates 检查是否存在新版本。
func (up Updater) CheckForUpdates(current *semver.Version, owner, repo string) (rel *Release, yes bool, err error) {
	if current == nil || owner == "" || repo == "" {
		return nil, false, fmt.Errorf("current version, owner, and repository are required")
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if !IsHttpSuccess(resp.StatusCode) {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
		return nil, false, fmt.Errorf("URL %q is unreachable", url)
	}

	var latest Release
	if err = decodeLimitedJSON(resp.Body, maxReleaseMetadataBytes, &latest); err != nil {
		return nil, false, err
	}

	latestVersion, err := semver.NewVersion(latest.TagName)
	if err != nil {
		return nil, false, err
	}
	if latestVersion.GreaterThan(current) {
		return &latest, true, nil
	}
	return nil, false, nil
}

// Apply 应用指定发布版本的更新。
func (up Updater) Apply(rel *Release,
	findAsset func([]Asset) (idx int),
	findChecksum func([]Asset) (algo Algorithm, expectedChecksum string, err error),
) error {
	if rel == nil || findAsset == nil || findChecksum == nil {
		return fmt.Errorf("update release or resolver is nil")
	}
	idx := findAsset(rel.Assets)
	if idx < 0 || idx >= len(rel.Assets) {
		return ErrAssetNotFound
	}

	algo, expectedChecksum, err := findChecksum(rel.Assets)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", strconv.FormatInt(time.Now().UnixNano(), 10))
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	downloadURL := rel.Assets[idx].BrowserDownloadURL
	fkDir := filepath.Join(tmpDir, "fkteams_update")
	downloadName := "fkteams_update.zip"
	if parsedURL, parseErr := url.Parse(downloadURL); parseErr == nil {
		if base := filepath.Base(parsedURL.Path); base != "." && base != ".." && base != string(filepath.Separator) && base != "" {
			downloadName = base
		}
	}
	srcFilename := filepath.Join(fkDir, downloadName)
	dstFilename := srcFilename

	// 配置HTTP客户端
	proxyStr := env.Get(env.ProxyURL)
	var proxyFunc func(*http.Request) (*url.URL, error)
	if proxyStr != "" {
		proxyURL, err := url.Parse(proxyStr)
		if err != nil {
			return fmt.Errorf("invalid proxy URL %q: %w", proxyStr, err)
		}
		proxyFunc = http.ProxyURL(proxyURL)
	} else {
		proxyFunc = http.ProxyFromEnvironment
	}
	transport := &http.Transport{
		Proxy:                 proxyFunc,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Minute,
	}

	// 设置进度回调
	var lastProgress float64
	progress := func(loaded, total int64, rate string) {
		if total <= 0 {
			return
		}
		progress := float64(loaded) / float64(total) * 100
		if progress > 100 {
			progress = 100
		}
		// 只在进度变化超过0.5%时更新显示
		if progress-lastProgress >= 0.5 || progress >= 100 {
			lastProgress = progress

			// 生成进度条
			barWidth := 40
			filledWidth := int(progress / 100 * float64(barWidth))
			bar := ""
			for i := range barWidth {
				if i < filledWidth {
					bar += "█"
				} else {
					bar += "░"
				}
			}

			// 显示进度
			fmt.Printf("\r[%s] %.2f%% | %s/%s | %s    ",
				bar, progress, formatFileSize(float64(loaded)), formatFileSize(float64(total)), rate)
		}
	}

	// 开始下载
	if err := downloadUpdateAsset(context.Background(), httpClient, downloadURL, dstFilename, maxUpdateArchiveBytes, progress); err != nil {
		fmt.Printf("下载失败: %v\n", err)
		return err
	}

	// 校验文件完整性
	fmt.Printf("\n基于 %s 校验文件完整性...\n", algo)
	if err = VerifyFile(algo, expectedChecksum, srcFilename); err != nil {
		return err
	}
	fmt.Printf("文件完整性校验通过\n")

	// 解压缩文件（如果需要）
	if rel.Assets[idx].IsCompressedFile() {
		if dstFilename, err = up.unarchive(srcFilename, tmpDir); err != nil {
			return err
		}
	}

	// 应用更新
	dstFile, err := os.Open(dstFilename)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	return selfupdate.Apply(dstFile, selfupdate.Options{})
}

func decodeLimitedJSON(reader io.Reader, limit int64, target any) error {
	data, err := readLimitedResponse(reader, limit)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode JSON response: %w", err)
	}
	return nil
}

func readLimitedResponse(reader io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("response exceeds %d bytes", limit)
	}
	return data, nil
}

func downloadUpdateAsset(
	ctx context.Context,
	client *http.Client,
	downloadURL string,
	destination string,
	limit int64,
	onProgress func(loaded, total int64, rate string),
) error {
	if client == nil {
		return fmt.Errorf("update HTTP client is nil")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("create update download request: %w", err)
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("download update: %w", err)
	}
	defer response.Body.Close()
	if !IsHttpSuccess(response.StatusCode) {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
		return fmt.Errorf("update download returned status %d", response.StatusCode)
	}
	if response.ContentLength > limit {
		return fmt.Errorf("update archive exceeds %d bytes", limit)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return fmt.Errorf("create update download directory: %w", err)
	}
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("create update archive: %w", err)
	}
	progressReader := &updateProgressReader{reader: io.LimitReader(response.Body, limit+1), total: response.ContentLength, callback: onProgress}
	written, copyErr := io.Copy(file, progressReader)
	if copyErr == nil && written > limit {
		copyErr = fmt.Errorf("update archive exceeds %d bytes", limit)
	}
	if copyErr == nil {
		copyErr = file.Sync()
	}
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(destination)
		return fmt.Errorf("write update archive: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(destination)
		return fmt.Errorf("close update archive: %w", closeErr)
	}
	return nil
}

type updateProgressReader struct {
	reader   io.Reader
	loaded   int64
	total    int64
	callback func(loaded, total int64, rate string)
}

func (reader *updateProgressReader) Read(buffer []byte) (int, error) {
	read, err := reader.reader.Read(buffer)
	reader.loaded += int64(read)
	if read > 0 && reader.callback != nil {
		reader.callback(reader.loaded, reader.total, "")
	}
	return read, err
}

// unarchive 解压文件并返回目标可执行文件路径。
func (up Updater) unarchive(srcFile, dstDir string) (dstFile string, err error) {
	if err = Unzip(srcFile, dstDir, func(processed, total int, fileName string, isDir bool) {
		fmt.Printf("解压中... %d/%d 文件: %s\n", processed, total, fileName)
	}); err != nil {
		return "", err
	}
	// locateTargetFile 在解压目录中查找可执行文件
	fis, err := os.ReadDir(dstDir)
	if err != nil {
		return "", fmt.Errorf("read extracted update directory: %w", err)
	}
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		name := fi.Name()
		if name == up.name || name == up.name+".exe" {
			return filepath.Join(dstDir, name), nil
		}
	}
	return "", fmt.Errorf("未在解压目录中找到可执行文件: %s", up.name)
}

// IsHttpSuccess 判断 HTTP 状态码是否表示成功。
func IsHttpSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}

// formatFileSize 将字节数格式化为可读字符串。
func formatFileSize(fileSize float64) string {
	const (
		KB = 1024.0
		MB = KB * 1024.0
		GB = MB * 1024.0
	)

	switch {
	case fileSize >= GB:
		return fmt.Sprintf("%.2f GB", fileSize/GB)
	case fileSize >= MB:
		return fmt.Sprintf("%.2f MB", fileSize/MB)
	case fileSize >= KB:
		return fmt.Sprintf("%.2f KB", fileSize/KB)
	default:
		return fmt.Sprintf("%.2f B", fileSize)
	}
}
