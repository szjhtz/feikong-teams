package handler

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

func TestArchiveProcessingWritesBoundedVisibleTree(t *testing.T) {
	filesystem := fstest.MapFS{
		"dir/a.txt":         {Data: []byte("a")},
		"dir/sub/b.txt":     {Data: []byte("content")},
		"dir/.hidden.txt":   {Data: []byte("hidden")},
		"dir/.hidden/value": {Data: []byte("hidden")},
	}
	sources := []archiveSource{{source: "dir", archiveBase: "bundle"}}
	if err := validateArchive(context.Background(), filesystem, sources); err != nil {
		t.Fatalf("validateArchive(): %v", err)
	}

	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	if err := writeArchive(context.Background(), filesystem, writer, sources); err != nil {
		t.Fatalf("writeArchive(): %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	reader, err := zip.NewReader(bytes.NewReader(output.Bytes()), int64(output.Len()))
	if err != nil {
		t.Fatal(err)
	}
	contents := make(map[string]string)
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		opened, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, readErr := io.ReadAll(opened)
		closeErr := opened.Close()
		if readErr != nil || closeErr != nil {
			t.Fatalf("read archive entry: %v, close: %v", readErr, closeErr)
		}
		contents[file.Name] = string(data)
	}
	if len(contents) != 2 || contents["bundle/a.txt"] != "a" || contents["bundle/sub/b.txt"] != "content" {
		t.Fatalf("archive contents = %#v", contents)
	}
}

func TestArchiveValidationHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := validateArchive(ctx, fstest.MapFS{"dir/file": {Data: []byte("data")}}, []archiveSource{{source: "dir"}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("validateArchive() error = %v, want canceled", err)
	}
}

func TestArchiveValidationLimitsEntryCount(t *testing.T) {
	filesystem := make(fstest.MapFS, maxArchiveEntries+1)
	for i := 0; i < maxArchiveEntries; i++ {
		filesystem[fmt.Sprintf("dir/%05d", i)] = &fstest.MapFile{}
	}
	err := validateArchive(context.Background(), filesystem, []archiveSource{{source: "dir", archiveBase: "dir"}})
	if !errors.Is(err, errArchiveLimit) {
		t.Fatalf("validateArchive() error = %v, want archive limit", err)
	}
}

func TestDownloadFileHandlerStreamsDirectoryArchive(t *testing.T) {
	workspace := setupWorkspaceDir(t)
	directory := filepath.Join(workspace, "docs")
	if err := os.Mkdir(directory, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "guide.txt"), []byte("guide"), 0644); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/download", DownloadFileHandler())
	response := performRequest(router, http.MethodGet, "/download?path=docs", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("download status = %d: %s", response.Code, response.Body.String())
	}
	reader, err := zip.NewReader(bytes.NewReader(response.Body.Bytes()), int64(response.Body.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if len(reader.File) != 1 || reader.File[0].Name != "guide.txt" {
		t.Fatalf("archive entries = %#v", reader.File)
	}
}

func TestBatchDownloadHandlerRejectsOverlappingPaths(t *testing.T) {
	workspace := setupWorkspaceDir(t)
	directory := filepath.Join(workspace, "docs")
	if err := os.Mkdir(directory, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "guide.txt"), []byte("guide"), 0644); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/download", BatchDownloadHandler())
	response := performJSON(router, http.MethodPost, "/download", `{"paths":["docs","docs/guide.txt"]}`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("batch download status = %d, want 400: %s", response.Code, response.Body.String())
	}
}
