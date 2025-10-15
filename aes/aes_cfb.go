package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// EncryptCFB 使用 AES-CFB 模式加密数据，key 必须是 16、24 或 32 字节
func EncryptCFB(plaintext, key []byte) (string, error) {
	// 创建加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 生成随机 IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// 创建 CFB 模式
	stream := cipher.NewCFBEncrypter(block, iv)

	// 加密数据
	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, plaintext)

	// 将 IV 和密文拼接并 Base64 编码
	result := append(iv, ciphertext...)
	return base64.StdEncoding.EncodeToString(result), nil
}

// DecryptCFB 使用 AES-CFB 模式解密数据，key 必须是 16、24 或 32 字节
func DecryptCFB(ciphertextBase64 string, key []byte) ([]byte, error) {
	// 解码 Base64
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, err
	}

	// 提取 IV
	if len(ciphertext) < aes.BlockSize {
		return nil, err
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// 创建解密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建 CFB 模式
	stream := cipher.NewCFBDecrypter(block, iv)

	// 解密数据
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}
