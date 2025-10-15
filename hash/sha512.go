package hash

import (
	"crypto/sha512"
	"encoding/hex"
)

// SHA512 计算给定数据的 SHA512 哈希值，返回 64 字节原始数据
func SHA512(data []byte) []byte {
	hash := sha512.Sum512(data)
	return hash[:]
}

// SHA512Hex 计算给定数据的 SHA512 哈希值，返回 128 位小写十六进制字符串
func SHA512Hex(data []byte) string {
	return hex.EncodeToString(SHA512(data))
}

// SHA384 计算给定数据的 SHA384 哈希值，返回 48 字节原始数据
func SHA384(data []byte) []byte {
	hash := sha512.Sum384(data)
	return hash[:]
}

// SHA384Hex 计算给定数据的 SHA384 哈希值，返回 96 位小写十六进制字符串
func SHA384Hex(data []byte) string {
	return hex.EncodeToString(SHA384(data))
}
