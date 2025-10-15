package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// EncryptGCM 使用 AES-GCM 模式加密数据，提供认证和加密
// key 必须是 16、24 或 32 字节（对应 AES-128/192/256）
// additionalData 可以为 nil
func EncryptGCM(plaintext, key, additionalData []byte) (string, error) {
	// 创建加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 创建 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密并认证数据
	ciphertext := gcm.Seal(nonce, nonce, plaintext, additionalData)

	// Base64 编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptGCM 使用 AES-GCM 模式解密数据，验证认证标签
// key 必须是 16、24 或 32 字节，data 为 Base64 编码字符串
// additionalData 必须与加密时相同
func DecryptGCM(ciphertextBase64 string, key, additionalData []byte) ([]byte, error) {
	// 解码 Base64
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, err
	}

	// 创建解密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 提取 nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密并验证
	plaintext, err := gcm.Open(nil, nonce, ciphertext, additionalData)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
