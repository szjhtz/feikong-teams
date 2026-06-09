package crypto

import (
	"crypto/aes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
)

var hexPattern = regexp.MustCompile(`^[0-9a-fA-F]{32}$`)

// EncryptAESECB 使用 AES-128-ECB 和 PKCS7 padding 加密明文。
func EncryptAESECB(plaintext, key []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, fmt.Errorf("AES key must be 16 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	padded := pkcs7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(padded))

	for i := 0; i < len(padded); i += aes.BlockSize {
		block.Encrypt(ciphertext[i:i+aes.BlockSize], padded[i:i+aes.BlockSize])
	}
	return ciphertext, nil
}

// DecryptAESECB 解密 AES-128-ECB 密文并移除 PKCS7 padding。
func DecryptAESECB(ciphertext, key []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, fmt.Errorf("AES key must be 16 bytes, got %d", len(key))
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext length %d is not a multiple of block size", len(ciphertext))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += aes.BlockSize {
		block.Decrypt(plaintext[i:i+aes.BlockSize], ciphertext[i:i+aes.BlockSize])
	}
	return pkcs7Unpad(plaintext)
}

// GenerateAESKey 生成 16 字节随机 AES key。
func GenerateAESKey() ([]byte, error) {
	key := make([]byte, 16)
	_, err := rand.Read(key)
	return key, err
}

// EncryptedSize 计算 AES-128-ECB 加密后的大小。
func EncryptedSize(rawSize int) int {
	return ((rawSize + 1 + aes.BlockSize - 1) / aes.BlockSize) * aes.BlockSize
}

// DecodeAESKey 解码协议中的 aes_key。
func DecodeAESKey(encoded string) ([]byte, error) {
	if hexPattern.MatchString(encoded) {
		return hex.DecodeString(encoded)
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("cannot base64 decode aes_key: %w", err)
		}
	}

	if len(decoded) == 16 {
		return decoded, nil
	}
	if len(decoded) == 32 && hexPattern.Match(decoded) {
		return hex.DecodeString(string(decoded))
	}

	return nil, fmt.Errorf("decoded aes_key has unexpected length %d (want 16 or 32)", len(decoded))
}

// EncodeAESKeyHex 返回 key 的十六进制字符串。
func EncodeAESKeyHex(key []byte) string {
	return hex.EncodeToString(key)
}

// EncodeAESKeyBase64 返回 CDNMedia.aes_key 使用的 base64(hex) 格式。
func EncodeAESKeyBase64(key []byte) string {
	return base64.StdEncoding.EncodeToString([]byte(hex.EncodeToString(key)))
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	pad := make([]byte, padding)
	for i := range pad {
		pad[i] = byte(padding)
	}
	return append(data, pad...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return nil, fmt.Errorf("invalid PKCS7 padding")
	}
	return data[:len(data)-padding], nil
}
