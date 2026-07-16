package update

import (
	"archive/zip"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestVerifyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payload.txt")
	content := []byte("fkteams update")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	sha256Sum := fmt.Sprintf("%x", sha256.Sum256(content))
	if err := VerifyFile(SHA256, sha256Sum, path); err != nil {
		t.Fatalf("VerifyFile SHA256 returned error: %v", err)
	}

	sha1Sum := fmt.Sprintf("%x", sha1.Sum(content))
	if err := VerifyFile(SHA1, sha1Sum, path); err != nil {
		t.Fatalf("VerifyFile SHA1 returned error: %v", err)
	}

	if err := VerifyFile(SHA256, "bad", path); err != ErrChecksumNotMatched {
		t.Fatalf("VerifyFile mismatch = %v, want ErrChecksumNotMatched", err)
	}
	if err := VerifyFile(Algorithm("MD5"), sha256Sum, path); err != ErrUnsupportedChecksumAlgorithm {
		t.Fatalf("VerifyFile unsupported = %v, want ErrUnsupportedChecksumAlgorithm", err)
	}
}

func TestFindAssetAndChecksum(t *testing.T) {
	suffix := fmt.Sprintf("%s_%s.zip", CapitalizeOS(), GetNormalizedArch())
	expected := strings.Repeat("a", 64)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s  fkteams_%s\n", expected, suffix)
	}))
	defer server.Close()

	items := []Asset{
		{Name: "other", BrowserDownloadURL: "https://example.com/fkteams_other.zip"},
		{Name: "binary", BrowserDownloadURL: "https://example.com/fkteams_" + suffix},
		{Name: "checksums.txt", BrowserDownloadURL: server.URL},
	}

	if idx := findAsset(items); idx != 1 {
		t.Fatalf("findAsset = %d, want 1", idx)
	}
	algo, got, err := findChecksum(items)
	if err != nil {
		t.Fatalf("findChecksum returned error: %v", err)
	}
	if algo != SHA256 || got != expected {
		t.Fatalf("findChecksum = (%s, %s), want (%s, %s)", algo, got, SHA256, expected)
	}
}

func TestFindChecksumMissingFile(t *testing.T) {
	_, _, err := findChecksum([]Asset{{Name: "binary", BrowserDownloadURL: "https://example.com/fkteams.zip"}})
	if err != ErrChecksumFileNotFound {
		t.Fatalf("findChecksum error = %v, want ErrChecksumFileNotFound", err)
	}
}

func TestAssetAndPlatformHelpers(t *testing.T) {
	if !(Asset{ContentType: "application/zip"}).IsCompressedFile() {
		t.Fatal("zip asset should be compressed")
	}
	if !(Asset{ContentType: "application/x-gzip"}).IsCompressedFile() {
		t.Fatal("gzip asset should be compressed")
	}
	if !(Asset{BrowserDownloadURL: "https://example.com/fkteams.zip"}).IsCompressedFile() {
		t.Fatal("zip URL should be compressed even without a content type")
	}
	if (Asset{ContentType: "application/octet-stream"}).IsCompressedFile() {
		t.Fatal("octet-stream asset should not be compressed")
	}

	if got := CapitalizeOS(); got == "" || got[0] < 'A' || got[0] > 'Z' {
		t.Fatalf("CapitalizeOS = %q, want capitalized value", got)
	}
	switch runtime.GOARCH {
	case "amd64":
		if got := GetNormalizedArch(); got != "x86_64" {
			t.Fatalf("GetNormalizedArch = %q, want x86_64", got)
		}
	case "386":
		if got := GetNormalizedArch(); got != "i386" {
			t.Fatalf("GetNormalizedArch = %q, want i386", got)
		}
	default:
		if got := GetNormalizedArch(); got != runtime.GOARCH {
			t.Fatalf("GetNormalizedArch = %q, want %q", got, runtime.GOARCH)
		}
	}

	if !IsHttpSuccess(http.StatusOK) || !IsHttpSuccess(http.StatusNoContent) {
		t.Fatal("2xx status should be successful")
	}
	if IsHttpSuccess(http.StatusMultipleChoices) || IsHttpSuccess(http.StatusInternalServerError) {
		t.Fatal("3xx/5xx status should not be successful")
	}
	if got := formatFileSize(1536); got != "1.50 KB" {
		t.Fatalf("formatFileSize = %q, want 1.50 KB", got)
	}
}

func TestUnzipExtractsFilesAndReportsProgress(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "archive.zip")
	createZip(t, zipPath, map[string]string{
		"bin/fkteams": "binary",
		"README.md":   "readme",
	})

	var seen []string
	dest := filepath.Join(tmp, "out")
	err := Unzip(zipPath, dest, func(processed int, total int, fileName string, isDir bool) {
		seen = append(seen, fmt.Sprintf("%d/%d:%s:%v", processed, total, fileName, isDir))
	})
	if err != nil {
		t.Fatalf("Unzip returned error: %v", err)
	}
	if len(seen) != 2 {
		t.Fatalf("progress callbacks = %#v, want two entries", seen)
	}
	if got := mustReadFile(t, filepath.Join(dest, "bin", "fkteams")); got != "binary" {
		t.Fatalf("extracted binary = %q, want binary", got)
	}
	if got := mustReadFile(t, filepath.Join(dest, "README.md")); got != "readme" {
		t.Fatalf("extracted README = %q, want readme", got)
	}
}

func TestUnzipRejectsZipSlip(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "slip.zip")
	createZip(t, zipPath, map[string]string{
		"../evil.txt": "evil",
	})

	err := Unzip(zipPath, filepath.Join(tmp, "out"), nil)
	if err == nil || !strings.Contains(err.Error(), "invalid update archive path") {
		t.Fatalf("Unzip error = %v, want Zip Slip rejection", err)
	}
	if _, statErr := os.Stat(filepath.Join(tmp, "evil.txt")); !os.IsNotExist(statErr) {
		t.Fatalf("evil file exists or unexpected stat error: %v", statErr)
	}
}

func TestUnzipRejectsExtractionLimit(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "large.zip")
	createZip(t, zipPath, map[string]string{"large.bin": "12345"})
	if err := unzipWithLimits(zipPath, filepath.Join(tmp, "out"), nil, 10, 4); err == nil {
		t.Fatal("oversized extracted update was accepted")
	}
}

func TestUnzipRejectsSymlink(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "symlink.zip")
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	header := &zip.FileHeader{Name: "link"}
	header.SetMode(os.ModeSymlink | 0777)
	entry, err := writer.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("target")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := Unzip(zipPath, filepath.Join(tmp, "out"), nil); err == nil {
		t.Fatal("symlink update archive entry was accepted")
	}
}

func TestReadLimitedResponseRejectsOversizedBody(t *testing.T) {
	if _, err := readLimitedResponse(strings.NewReader("12345"), 4); err == nil {
		t.Fatal("oversized update response was accepted")
	}
}

func TestDownloadUpdateAssetRejectsOversizedBody(t *testing.T) {
	client := &http.Client{Transport: updateRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    http.StatusOK,
			ContentLength: -1,
			Header:        make(http.Header),
			Body:          io.NopCloser(strings.NewReader("12345")),
			Request:       request,
		}, nil
	})}
	destination := filepath.Join(t.TempDir(), "update.zip")
	if err := downloadUpdateAsset(context.Background(), client, "https://example.com/update.zip", destination, 4, nil); err == nil {
		t.Fatal("oversized update download was accepted")
	}
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("rejected update archive was not removed: %v", err)
	}
}

type updateRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn updateRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func createZip(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create zip entry %s: %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry %s: %v", name, err)
		}
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
