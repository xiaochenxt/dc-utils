package hash

import (
	"crypto/sha1"
	"encoding/hex"
)

// SHA1 计算给定数据的 SHA1 哈希值，返回 20 字节原始数据
func SHA1(data []byte) []byte {
	hash := sha1.Sum(data)
	return hash[:]
}

// SHA1Hex 计算给定数据的 SHA1 哈希值，返回 40 位小写十六进制字符串
func SHA1Hex(data []byte) string {
	return hex.EncodeToString(SHA1(data))
}
