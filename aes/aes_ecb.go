package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

// EncryptECB 使用 AES-ECB 模式加密数据，key 必须是 16、24 或 32 字节（对应 AES-128/192/256）
// 注意：ECB模式不安全，建议仅用于短数据或测试
func EncryptECB(plaintext, key []byte) (string, error) {
	// 创建加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 填充数据到块大小的整数倍
	plaintext = pkcs7Padding(plaintext, block.BlockSize())

	// 创建 ECB 模式
	mode := newECBEncrypter(block)

	// 加密数据
	ciphertext := make([]byte, len(plaintext))
	mode.CryptBlocks(ciphertext, plaintext)

	// Base64 编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptECB 使用 AES-ECB 模式解密数据，key 必须是 16、24 或 32 字节，data 为 Base64 编码字符串
// 注意：ECB模式不安全，建议仅用于短数据或测试
func DecryptECB(ciphertextBase64 string, key []byte) ([]byte, error) {
	// 解码 Base64
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, err
	}

	// 检查数据长度
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	// 创建解密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建 ECB 模式
	mode := newECBDecrypter(block)

	// 解密数据
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// 去除填充
	plaintext = pkcs7UnPadding(plaintext)

	return plaintext, nil
}

// ecbEncrypter 实现 ECB 加密模式
type ecbEncrypter struct {
	b         cipher.Block
	blockSize int
}

func newECBEncrypter(b cipher.Block) cipher.BlockMode {
	return &ecbEncrypter{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

func (x *ecbEncrypter) BlockSize() int { return x.blockSize }

func (x *ecbEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for i := 0; i < len(src); i += x.blockSize {
		x.b.Encrypt(dst[i:i+x.blockSize], src[i:i+x.blockSize])
	}
}

// ecbDecrypter 实现 ECB 解密模式
type ecbDecrypter struct {
	b         cipher.Block
	blockSize int
}

func newECBDecrypter(b cipher.Block) cipher.BlockMode {
	return &ecbDecrypter{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

func (x *ecbDecrypter) BlockSize() int { return x.blockSize }

func (x *ecbDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for i := 0; i < len(src); i += x.blockSize {
		x.b.Decrypt(dst[i:i+x.blockSize], src[i:i+x.blockSize])
	}
}
