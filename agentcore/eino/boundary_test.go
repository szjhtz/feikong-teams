package eino

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func TestEinoImportsStayInsideAdapter(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	adapterRoot := filepath.Join("agentcore", "eino")
	einoPrefix := "github.com/cloudwego/" + "eino"

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "release", "node_modules", "web":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			importPath := strings.Trim(spec.Path.Value, `"`)
			if strings.HasPrefix(importPath, einoPrefix) && !strings.HasPrefix(rel, filepath.ToSlash(adapterRoot)+"/") {
				t.Errorf("%s imports %s outside %s", rel, importPath, filepath.ToSlash(adapterRoot))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
