//go:build windows

package atomicfile

import "os"

// Windows 不提供可移植的目录句柄同步语义，文件内容仍会在替换前完成同步。
func syncDirectoryFile(_ *os.File) error {
	return nil
}
