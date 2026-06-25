package update

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"os"
)

// Algorithm 表示校验和算法。
type Algorithm string

const (
	// SHA256 使用 SHA-256。
	SHA256 Algorithm = "SHA256"
	// SHA1 使用 SHA-1。
	SHA1 Algorithm = "SHA1"
)

// VerifyFile 根据期望校验和验证文件完整性。
func VerifyFile(algo Algorithm, expectedChecksum, filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	var h hash.Hash
	switch algo {
	case SHA256:
		h = sha256.New()
	case SHA1:
		h = sha1.New()
	default:
		return ErrUnsupportedChecksumAlgorithm
	}

	if _, err = io.Copy(h, f); err != nil {
		return err
	}

	if expectedChecksum != hex.EncodeToString(h.Sum(nil)) {
		return ErrChecksumNotMatched
	}
	return nil
}
