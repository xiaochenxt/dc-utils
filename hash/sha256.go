package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256 计算给定数据的 SHA256 哈希值，返回 32 字节原始数据
func SHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// SHA256Hex 计算给定数据的 SHA256 哈希值，返回 64 位小写十六进制字符串
func SHA256Hex(data []byte) string {
	return hex.EncodeToString(SHA256(data))
}
