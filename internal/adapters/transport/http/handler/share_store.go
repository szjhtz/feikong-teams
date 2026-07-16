package handler

import (
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	maxPersistedShareStoreBytes = 16 << 20
	maxPersistedShareEntries    = 10_000
)

var errShareStoreFull = errors.New("share store entry limit reached")

func readPersistedShareStore(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxPersistedShareStoreBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxPersistedShareStoreBytes {
		return nil, fmt.Errorf("share store file is too large")
	}
	return data, nil
}

func validatePersistedShareStoreSize(data []byte) error {
	if len(data) > maxPersistedShareStoreBytes {
		return fmt.Errorf("share store file is too large")
	}
	return nil
}
