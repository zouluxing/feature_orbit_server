package crypto_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zouluxing/ums/pkg/crypto"
)

// validKey is a 32-byte key represented as 64 hex chars.
const validKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestNewAESEncryptor_ValidKey(t *testing.T) {
	_, err := crypto.NewAESEncryptor(validKey)
	require.NoError(t, err)
}

func TestNewAESEncryptor_EmptyKey(t *testing.T) {
	_, err := crypto.NewAESEncryptor("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not set")
}

func TestNewAESEncryptor_ShortKey(t *testing.T) {
	_, err := crypto.NewAESEncryptor("deadbeef")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "32 bytes")
}

func TestNewAESEncryptor_InvalidHex(t *testing.T) {
	_, err := crypto.NewAESEncryptor(strings.Repeat("zz", 32))
	assert.Error(t, err)
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	enc, err := crypto.NewAESEncryptor(validKey)
	require.NoError(t, err)

	for _, plain := range []string{
		"JBSWY3DPEHPK3PXP",          // typical TOTP secret
		"hello world",               // short
		strings.Repeat("x", 512),    // long
		"",                          // empty string
		"特殊字符 \u0000 \n \t",      // unicode + control chars
	} {
		cipher, err := enc.Encrypt(plain)
		require.NoError(t, err, "Encrypt(%q)", plain)
		assert.NotEqual(t, plain, cipher, "ciphertext must differ from plaintext")

		got, err := enc.Decrypt(cipher)
		require.NoError(t, err, "Decrypt round-trip for %q", plain)
		assert.Equal(t, plain, got)
	}
}

func TestEncrypt_NonDeterministic(t *testing.T) {
	enc, _ := crypto.NewAESEncryptor(validKey)
	a, _ := enc.Encrypt("same plaintext")
	b, _ := enc.Encrypt("same plaintext")
	assert.NotEqual(t, a, b, "each encryption must use a fresh random nonce")
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	enc, _ := crypto.NewAESEncryptor(validKey)
	cipher, _ := enc.Encrypt("sensitive-totp-secret")
	// Flip the last byte of the base64 payload
	tampered := cipher[:len(cipher)-1] + "X"
	_, err := enc.Decrypt(tampered)
	assert.Error(t, err, "tampered ciphertext must not decrypt")
}

func TestDecrypt_TooShort(t *testing.T) {
	enc, _ := crypto.NewAESEncryptor(validKey)
	_, err := enc.Decrypt("abc")
	assert.Error(t, err)
}

func TestDecrypt_WrongKey(t *testing.T) {
	enc1, _ := crypto.NewAESEncryptor(validKey)
	enc2, _ := crypto.NewAESEncryptor(strings.Repeat("ff", 32))

	cipher, _ := enc1.Encrypt("secret")
	_, err := enc2.Decrypt(cipher)
	assert.Error(t, err, "decryption with wrong key must fail")
}
