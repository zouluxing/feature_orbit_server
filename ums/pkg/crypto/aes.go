// Package crypto provides AES-256-GCM authenticated encryption utilities
// for protecting sensitive fields (e.g. MFA TOTP secrets) at rest.
//
// Usage:
//
//	enc, err := crypto.NewAESEncryptor(os.Getenv("UMS_MFA_ENCRYPTION_KEY"))
//	ciphertext, err := enc.Encrypt("JBSWY3DPEHPK3PXP")  // TOTP secret
//	plaintext, err := enc.Decrypt(ciphertext)
//
// The encryption key must be exactly 32 bytes (256-bit), supplied as a
// hex-encoded string via the UMS_MFA_ENCRYPTION_KEY environment variable.
// Generate a key with:  openssl rand -hex 32
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// AESEncryptor performs AES-256-GCM authenticated encryption / decryption.
type AESEncryptor struct {
	key []byte // 32-byte AES-256 key
}

// NewAESEncryptor creates an AESEncryptor from a 64-character hex-encoded key.
// Returns an error if the key is missing, malformed, or not exactly 32 bytes.
func NewAESEncryptor(hexKey string) (*AESEncryptor, error) {
	if hexKey == "" {
		return nil, errors.New("crypto: UMS_MFA_ENCRYPTION_KEY is not set")
	}
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("crypto: invalid hex key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("crypto: key must be 32 bytes (64 hex chars), got %d bytes", len(key))
	}
	return &AESEncryptor{key: key}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns a base64url-encoded
// string in the format:  nonce(12B) || ciphertext || tag(16B)
func (e *AESEncryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("crypto: create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("crypto: create GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize()) // 12 bytes
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("crypto: generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64url-encoded ciphertext produced by Encrypt.
func (e *AESEncryptor) Decrypt(encoded string) (string, error) {
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("crypto: base64 decode: %w", err)
	}
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("crypto: create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("crypto: create GCM: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize+gcm.Overhead() {
		return "", errors.New("crypto: ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("crypto: decrypt/authenticate: %w", err)
	}
	return string(plaintext), nil
}
