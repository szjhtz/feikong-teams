package web

import (
	"embed"
	"io/fs"
)

// FS 嵌入前端生产构建产物。
//
//go:embed dist
var FS embed.FS

// GetFS 返回前端生产构建目录。
func GetFS() fs.FS {
	sub, err := fs.Sub(FS, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}
