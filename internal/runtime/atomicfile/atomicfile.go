package atomicfile

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// WriteFunc 将目标内容写入临时文件。返回值是已写入字节数。
type WriteFunc func(io.Writer) (int64, error)

// WriteFile 将 data 持久化到同目录临时文件后原子替换目标文件。
func WriteFile(path string, data []byte, perm os.FileMode) error {
	directory := filepath.Dir(path)
	existingAncestor, err := ensureParentDirectory(directory)
	if err != nil {
		return err
	}

	root, err := os.OpenRoot(directory)
	if err != nil {
		return fmt.Errorf("open target directory: %w", err)
	}
	_, writeErr := WriteInRoot(root, filepath.Base(path), perm, func(writer io.Writer) (int64, error) {
		written, err := writer.Write(data)
		if err == nil && written != len(data) {
			err = io.ErrShortWrite
		}
		return int64(written), err
	})
	closeErr := root.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return fmt.Errorf("close target directory: %w", closeErr)
	}

	// 新建目录本身也是父目录中的目录项，需要逐层同步到最近的既有祖先。
	if directory != existingAncestor {
		for current := filepath.Dir(directory); ; current = filepath.Dir(current) {
			if err := syncDirectory(current); err != nil {
				return fmt.Errorf("sync parent directory: %w", err)
			}
			if current == existingAncestor {
				break
			}
		}
	}
	return nil
}

// WriteFileInRoot 在 root 内原子替换相对路径文件。目标父目录必须已存在。
func WriteFileInRoot(root *os.Root, path string, data []byte, perm os.FileMode) error {
	_, err := WriteInRoot(root, path, perm, func(writer io.Writer) (int64, error) {
		written, err := writer.Write(data)
		if err == nil && written != len(data) {
			err = io.ErrShortWrite
		}
		return int64(written), err
	})
	return err
}

// WriteReaderInRoot 将 reader 的有限内容原子写入 root 内的相对路径。
func WriteReaderInRoot(root *os.Root, path string, reader io.Reader, maxBytes int64, perm os.FileMode) (int64, error) {
	if maxBytes < 0 {
		return 0, fmt.Errorf("max bytes must not be negative")
	}
	return WriteInRoot(root, path, perm, func(writer io.Writer) (int64, error) {
		written, err := io.Copy(writer, io.LimitReader(reader, maxBytes+1))
		if err != nil {
			return written, err
		}
		if written > maxBytes {
			return written, fmt.Errorf("file exceeds size limit")
		}
		return written, nil
	})
}

// WriteInRoot 在 root 内的目标同目录创建临时文件，完整同步后再原子替换。
func WriteInRoot(root *os.Root, path string, perm os.FileMode, write WriteFunc) (int64, error) {
	if root == nil {
		return 0, fmt.Errorf("root is nil")
	}
	if write == nil {
		return 0, fmt.Errorf("write function is nil")
	}
	cleanPath := filepath.Clean(path)
	if cleanPath == "." || filepath.Base(cleanPath) == "." {
		return 0, fmt.Errorf("target path is invalid")
	}

	temporaryPath, temporary, err := createTemporaryFile(root, filepath.Dir(cleanPath))
	if err != nil {
		return 0, err
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = root.Remove(temporaryPath)
		}
	}()

	written, writeErr := write(temporary)
	if writeErr != nil {
		_ = temporary.Close()
		return written, fmt.Errorf("write temp file: %w", writeErr)
	}
	if err := temporary.Chmod(perm.Perm()); err != nil {
		_ = temporary.Close()
		return written, fmt.Errorf("chmod temp file: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return written, fmt.Errorf("sync temp file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return written, fmt.Errorf("close temp file: %w", err)
	}
	if err := root.Rename(temporaryPath, cleanPath); err != nil {
		return written, fmt.Errorf("replace file: %w", err)
	}
	cleanup = false
	if err := syncRootDirectory(root, filepath.Dir(cleanPath)); err != nil {
		return written, fmt.Errorf("sync target directory: %w", err)
	}
	return written, nil
}

func ensureParentDirectory(directory string) (string, error) {
	existing := filepath.Clean(directory)
	for {
		_, err := os.Stat(existing)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("inspect parent directory: %w", err)
		}
		parent := filepath.Dir(existing)
		if parent == existing {
			return "", fmt.Errorf("find existing parent directory: %w", err)
		}
		existing = parent
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		return "", fmt.Errorf("create parent directory: %w", err)
	}
	return existing, nil
}

func createTemporaryFile(root *os.Root, directory string) (string, *os.File, error) {
	for range 10 {
		var random [16]byte
		if _, err := io.ReadFull(rand.Reader, random[:]); err != nil {
			return "", nil, fmt.Errorf("generate temp file name: %w", err)
		}
		name := ".fkteams-tmp-" + hex.EncodeToString(random[:])
		path := name
		if directory != "." {
			path = filepath.Join(directory, name)
		}
		file, err := root.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err == nil {
			return path, file, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return "", nil, fmt.Errorf("create temp file: %w", err)
		}
	}
	return "", nil, fmt.Errorf("create temp file: too many name collisions")
}

func syncRootDirectory(root *os.Root, directory string) error {
	opened, err := root.Open(directory)
	if err != nil {
		return err
	}
	syncErr := syncDirectoryFile(opened)
	closeErr := opened.Close()
	if syncErr != nil {
		return syncErr
	}
	return closeErr
}

func syncDirectory(path string) error {
	opened, err := os.Open(path)
	if err != nil {
		return err
	}
	syncErr := syncDirectoryFile(opened)
	closeErr := opened.Close()
	if syncErr != nil {
		return syncErr
	}
	return closeErr
}
