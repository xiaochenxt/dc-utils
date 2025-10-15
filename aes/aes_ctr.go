package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// EncryptCTR 使用 AES-CTR 模式加密数据，key 必须是 16、24 或 32 字节
func EncryptCTR(plaintext, key []byte) (string, error) {
	// 创建加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 生成随机 nonce（用作计数器初始值）
	nonce := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 创建 CTR 模式
	stream := cipher.NewCTR(block, nonce)

	// 加密数据
	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, plaintext)

	// 将 nonce 和密文拼接并 Base64 编码
	result := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(result), nil
}

// DecryptCTR 使用 AES-CTR 模式解密数据，key 必须是 16、24 或 32 字节
func DecryptCTR(ciphertextBase64 string, key []byte) ([]byte, error) {
	// 解码 Base64
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, err
	}

	// 提取 nonce
	if len(ciphertext) < aes.BlockSize {
		return nil, err
	}
	nonce := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// 创建解密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建 CTR 模式
	stream := cipher.NewCTR(block, nonce)

	// 解密数据
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}
