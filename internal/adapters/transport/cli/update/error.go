package update

import "errors"

var (
	// ErrUnsupportedChecksumAlgorithm 表示不支持的校验和算法。
	ErrUnsupportedChecksumAlgorithm = errors.New("unsupported checksum algorithm")
	// ErrChecksumNotMatched 表示文件校验和不匹配。
	ErrChecksumNotMatched = errors.New("file checksum does not match the computed checksum")
	// ErrChecksumFileNotFound 表示未找到校验和文件。
	ErrChecksumFileNotFound = errors.New("checksum file not found")
	// ErrAssetNotFound 表示未找到发布资源。
	ErrAssetNotFound = errors.New("asset not found")
	// ErrCollectorNotFound 表示未找到收集器。
	ErrCollectorNotFound = errors.New("collector not found")
	// ErrEmptyURL 表示 URL 为空。
	ErrEmptyURL = errors.New("empty url")
)
