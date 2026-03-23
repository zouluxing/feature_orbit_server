package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UMSClaims struct {
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	Scopes []string `json:"scopes,omitempty"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	keyID      string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTManager(privPath, pubPath, issuer, keyID string, accessTTL, refreshTTL time.Duration) (*JWTManager, error) {
	privPEM, err := os.ReadFile(privPath)
	if err != nil { return nil, fmt.Errorf("read private key: %w", err) }
	block, _ := pem.Decode(privPEM)
	if block == nil { return nil, fmt.Errorf("invalid PEM block for private key") }
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		privKeyAny, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil { return nil, fmt.Errorf("parse private key: %w", err) }
		var ok bool
		privKey, ok = privKeyAny.(*rsa.PrivateKey)
		if !ok { return nil, fmt.Errorf("private key is not RSA") }
	}
	pubPEM, err := os.ReadFile(pubPath)
	if err != nil { return nil, fmt.Errorf("read public key: %w", err) }
	block, _ = pem.Decode(pubPEM)
	if block == nil { return nil, fmt.Errorf("invalid PEM block for public key") }
	pubKeyAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil { return nil, fmt.Errorf("parse public key: %w", err) }
	pubKey, ok := pubKeyAny.(*rsa.PublicKey)
	if !ok { return nil, fmt.Errorf("public key is not RSA") }
	return &JWTManager{privateKey: privKey, publicKey: pubKey, issuer: issuer, keyID: keyID, accessTTL: accessTTL, refreshTTL: refreshTTL}, nil
}

func (m *JWTManager) IssueAccessToken(userUUID, email string, roles, scopes []string) (string, string, error) {
	jti := NewUUID(); now := time.Now()
	claims := UMSClaims{
		Email: email, Roles: roles, Scopes: scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: m.issuer, Subject: userUUID,
			IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)), ID: jti,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = m.keyID
	signed, err := token.SignedString(m.privateKey)
	return signed, jti, err
}

func (m *JWTManager) VerifyAccessToken(tokenStr string) (*UMSClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UMSClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok { return nil, fmt.Errorf("unexpected alg: %v", t.Header["alg"]) }
		return m.publicKey, nil
	})
	if err != nil { return nil, err }
	claims, ok := token.Claims.(*UMSClaims)
	if !ok || !token.Valid { return nil, fmt.Errorf("invalid token claims") }
	return claims, nil
}

func (m *JWTManager) PublicKeyN() string { return base64.RawURLEncoding.EncodeToString(m.publicKey.N.Bytes()) }
func (m *JWTManager) PublicKeyE() string { e := big.NewInt(int64(m.publicKey.E)); return base64.RawURLEncoding.EncodeToString(e.Bytes()) }
func (m *JWTManager) KeyID() string            { return m.keyID }
func (m *JWTManager) AccessTTL() time.Duration  { return m.accessTTL }

func GenerateRSAKeyPairFiles(privPath, pubPath string) error {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil { return err }
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privKey)})
	if err := os.WriteFile(privPath, privPEM, 0600); err != nil { return err }
	pubDER, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil { return err }
	return os.WriteFile(pubPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}), 0644)
}
