package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// EncryptCBC 使用 AES-CBC 模式加密数据，key 必须是 16、24 或 32 字节（对应 AES-128/192/256）
func EncryptCBC(plaintext, key []byte) (string, error) {
	// 生成随机 IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// 创建加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 填充数据到块大小的整数倍
	plaintext = pkcs7Padding(plaintext, block.BlockSize())

	// 创建 CBC 模式
	mode := cipher.NewCBCEncrypter(block, iv)

	// 加密数据
	ciphertext := make([]byte, len(plaintext))
	mode.CryptBlocks(ciphertext, plaintext)

	// 将 IV 和密文拼接并 Base64 编码
	result := append(iv, ciphertext...)
	return base64.StdEncoding.EncodeToString(result), nil
}

// DecryptCBC 使用 AES-CBC 模式解密数据，key 必须是 16、24 或 32 字节，data 为 Base64 编码字符串
func DecryptCBC(ciphertextBase64 string, key []byte) ([]byte, error) {
	// 解码 Base64
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, err
	}

	// 检查数据长度
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	// 提取 IV 和密文
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// 创建解密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建 CBC 模式
	mode := cipher.NewCBCDecrypter(block, iv)

	// 解密数据
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// 去除填充
	plaintext = pkcs7UnPadding(plaintext)

	return plaintext, nil
}

// pkcs7Padding 实现 PKCS#7 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// pkcs7UnPadding 移除 PKCS#7 填充
func pkcs7UnPadding(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return nil
	}
	unpadding := int(data[length-1])
	if unpadding > length || unpadding == 0 {
		return nil
	}
	for i := 0; i < unpadding; i++ {
		if data[length-unpadding+i] != byte(unpadding) {
			return nil
		}
	}
	return data[:length-unpadding]
}
