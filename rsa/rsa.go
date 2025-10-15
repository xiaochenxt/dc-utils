package rsa

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
)

// GenerateKeyPair 生成 RSA 密钥对，bits 为密钥长度（如 2048）
func GenerateKeyPair(bits int) (privateKey, publicKey []byte, err error) {
	private, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	// 生成私钥 PEM 格式
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(private),
	})

	// 生成公钥 PEM 格式
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&private.PublicKey),
	})

	return privateKeyPEM, publicKeyPEM, nil
}

// EncryptOAEP 使用 RSA-OAEP 加密数据，返回 Base64 编码字符串
func EncryptOAEP(plaintext []byte, publicKeyPEM []byte) (string, error) {
	// 解析 PEM 格式的公钥
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return "", errors.New("failed to parse public key PEM")
	}

	pubInterface, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return "", err
	}

	// 加密数据
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubInterface, plaintext, nil)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptOAEP 使用 RSA-OAEP 解密数据，data 为 Base64 编码字符串
func DecryptOAEP(ciphertextBase64 string, privateKeyPEM []byte) ([]byte, error) {
	// 解析 PEM 格式的私钥
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, errors.New("failed to parse private key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// 解码 Base64
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, err
	}

	// 解密数据
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// SignPKCS1v15 使用 RSA-PKCS1v15 进行签名，返回 Base64 编码字符串
func SignPKCS1v15(data []byte, privateKeyPEM []byte) (string, error) {
	// 解析 PEM 格式的私钥
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return "", errors.New("failed to parse private key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	// 计算哈希
	hashed := sha256.Sum256(data)

	// 签名
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifyPKCS1v15 使用 RSA-PKCS1v15 验证签名，signature 为 Base64 编码字符串
func VerifyPKCS1v15(data []byte, signatureBase64 string, publicKeyPEM []byte) error {
	// 解析 PEM 格式的公钥
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return errors.New("failed to parse public key PEM")
	}

	pubInterface, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return err
	}

	// 解码 Base64 签名
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return err
	}

	// 计算哈希
	hashed := sha256.Sum256(data)

	// 验证签名
	return rsa.VerifyPKCS1v15(pubInterface, crypto.SHA256, hashed[:], signature)
}
