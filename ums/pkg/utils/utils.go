package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"math/big"
	"strings"

	"github.com/google/uuid"
)

func NewUUID() string { return uuid.New().String() }

func RandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil { panic(err) }
	return base64.RawURLEncoding.EncodeToString(b)
}

func RandomCode(n int) string {
	const digits = "0123456789"
	sb := strings.Builder{}
	for i := 0; i < n; i++ {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		sb.WriteByte(digits[idx.Int64()])
	}
	return sb.String()
}

func SHA256Base64(s string) string {
	h := sha256.Sum256([]byte(s))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func DefaultInt(val, def int) int {
	if val <= 0 { return def }
	return val
}

func Offset(page, pageSize int) int {
	if page < 1 { page = 1 }
	return (page - 1) * pageSize
}
