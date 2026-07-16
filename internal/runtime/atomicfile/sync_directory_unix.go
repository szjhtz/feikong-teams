//go:build !windows

package atomicfile

import "os"

func syncDirectoryFile(directory *os.File) error {
	return directory.Sync()
}
