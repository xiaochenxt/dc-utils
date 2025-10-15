package hash

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5 计算给定数据的 MD5 哈希值，返回 16 字节原始数据
func MD5(data []byte) []byte {
	hash := md5.Sum(data)
	return hash[:]
}

// MD5Hex 计算给定数据的 MD5 哈希值，返回 32 位小写十六进制字符串
func MD5Hex(data []byte) string {
	return hex.EncodeToString(MD5(data))
}
