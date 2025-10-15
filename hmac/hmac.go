package hmac

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
)

// HMACSHA1 计算给定数据和密钥的 HMAC-SHA1 值，返回 20 字节原始数据
func HMACSHA1(data, key []byte) []byte {
	h := hmac.New(sha1.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// HMACSHA1Hex 计算给定数据和密钥的 HMAC-SHA1 值，返回 40 位小写十六进制字符串
func HMACSHA1Hex(data, key []byte) string {
	return hex.EncodeToString(HMACSHA1(data, key))
}

// HMACSHA256 计算给定数据和密钥的 HMAC-SHA256 值，返回 32 字节原始数据
func HMACSHA256(data, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// HMACSHA256Hex 计算给定数据和密钥的 HMAC-SHA256 值，返回 64 位小写十六进制字符串
func HMACSHA256Hex(data, key []byte) string {
	return hex.EncodeToString(HMACSHA256(data, key))
}

// HMACSHA512 计算给定数据和密钥的 HMAC-SHA512 值，返回 64 字节原始数据
func HMACSHA512(data, key []byte) []byte {
	h := hmac.New(sha512.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// HMACSHA512Hex 计算给定数据和密钥的 HMAC-SHA512 值，返回 128 位小写十六进制字符串
func HMACSHA512Hex(data, key []byte) string {
	return hex.EncodeToString(HMACSHA512(data, key))
}

// HMACSHA384 计算给定数据和密钥的 HMAC-SHA384 值，返回 48 字节原始数据
func HMACSHA384(data, key []byte) []byte {
	h := hmac.New(sha512.New384, key)
	h.Write(data)
	return h.Sum(nil)
}

// HMACSHA384Hex 计算给定数据和密钥的 HMAC-SHA384 值，返回 96 位小写十六进制字符串
func HMACSHA384Hex(data, key []byte) string {
	return hex.EncodeToString(HMACSHA384(data, key))
}
