package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
)

// Crypto AES-GCM 加密服务
type Crypto struct {
	key []byte
}

// NewCrypto 创建加密服务实例
// key 必须是 16, 24, 或 32 字节（对应 AES-128, AES-192, AES-256）
func NewCrypto(key string) (*Crypto, error) {
	keyBytes := []byte(key)
	keyLen := len(keyBytes)

	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, errors.New("加密密钥长度必须是 16, 24 或 32 字节")
	}

	return &Crypto{key: keyBytes}, nil
}

// Encrypt 加密字符串，返回 base64 编码的密文
func (c *Crypto) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密，nonce 附加在密文前面
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 base64 编码的密文
func (c *Crypto) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("密文长度不足")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// MaskPhone 手机号脱敏显示（138****8888）
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// Hash 计算字符串的 SHA256 哈希（用于索引查询）
func Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
