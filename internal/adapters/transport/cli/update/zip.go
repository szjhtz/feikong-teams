package update

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxUpdateArchiveEntries       = 1_000
	maxUpdateExtractedBytes int64 = 1 << 30
)

// UnzipCallback 定义进度回调函数
type UnzipCallback func(processed int, total int, fileName string, isDir bool)

// Unzip 带进度回调并在固定资源边界内解压更新包。
func Unzip(source, destination string, callback UnzipCallback) error {
	return unzipWithLimits(source, destination, callback, maxUpdateArchiveEntries, maxUpdateExtractedBytes)
}

func unzipWithLimits(source, destination string, callback UnzipCallback, maxEntries int, maxBytes int64) error {
	if maxEntries < 0 || maxBytes < 0 {
		return fmt.Errorf("update archive limits must not be negative")
	}
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()
	if len(reader.File) > maxEntries {
		return fmt.Errorf("update archive exceeds %d entries", maxEntries)
	}

	destDir := filepath.Clean(destination)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(reader.File))
	var extractedBytes int64
	for index, file := range reader.File {
		targetPath, err := safeUpdateArchiveTarget(destDir, file.Name)
		if err != nil {
			return err
		}
		if _, exists := seen[targetPath]; exists {
			return fmt.Errorf("duplicate update archive path: %s", file.Name)
		}
		seen[targetPath] = struct{}{}
		mode := file.Mode()
		if mode&os.ModeSymlink != 0 || (!file.FileInfo().IsDir() && !mode.IsRegular()) {
			return fmt.Errorf("unsupported update archive entry: %s", file.Name)
		}
		if callback != nil {
			callback(index+1, len(reader.File), file.Name, file.FileInfo().IsDir())
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
			continue
		}
		if file.UncompressedSize64 > uint64(maxBytes-extractedBytes) {
			return fmt.Errorf("extracted update exceeds %d bytes", maxBytes)
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		written, err := extractUpdateFile(file, targetPath, maxBytes-extractedBytes)
		if err != nil {
			return fmt.Errorf("extract update file %s: %w", file.Name, err)
		}
		extractedBytes += written
	}
	return nil
}

func safeUpdateArchiveTarget(destination, archivePath string) (string, error) {
	if archivePath == "" || strings.ContainsRune(archivePath, '\x00') {
		return "", fmt.Errorf("invalid update archive path")
	}
	cleanRelative := filepath.Clean(filepath.FromSlash(archivePath))
	if cleanRelative == "." || filepath.IsAbs(cleanRelative) || filepath.VolumeName(cleanRelative) != "" {
		return "", fmt.Errorf("invalid update archive path: %s", archivePath)
	}
	targetPath := filepath.Join(destination, cleanRelative)
	relative, err := filepath.Rel(destination, targetPath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("invalid update archive path: %s", archivePath)
	}
	return targetPath, nil
}

func extractUpdateFile(file *zip.File, targetPath string, limit int64) (int64, error) {
	reader, err := file.Open()
	if err != nil {
		return 0, err
	}
	permissions := os.FileMode(0644)
	if file.Mode().Perm()&0111 != 0 {
		permissions = 0755
	}
	target, err := os.OpenFile(targetPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, permissions)
	if err != nil {
		_ = reader.Close()
		return 0, err
	}
	written, copyErr := io.Copy(target, io.LimitReader(reader, limit+1))
	if copyErr == nil && written > limit {
		copyErr = fmt.Errorf("extracted update exceeds size limit")
	}
	if copyErr == nil {
		copyErr = target.Sync()
	}
	readerCloseErr := reader.Close()
	targetCloseErr := target.Close()
	if copyErr != nil {
		return written, copyErr
	}
	if readerCloseErr != nil {
		return written, readerCloseErr
	}
	if targetCloseErr != nil {
		return written, targetCloseErr
	}
	return written, nil
}
