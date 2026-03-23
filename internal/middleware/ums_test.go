package middleware_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zouluxing/feature_orbit_server/internal/middleware"
)

// ── Test helpers ─────────────────────────────────────────────────────────────

func generateTestKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

func jwksServer(t *testing.T, key *rsa.PrivateKey, kid string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nB := base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes())
		eB := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"keys": []map[string]string{
				{"kty": "RSA", "kid": kid, "n": nB, "e": eB},
			},
		})
	}))
}

func signToken(t *testing.T, key *rsa.PrivateKey, kid, subject, email string, roles []string, ttl time.Duration) string {
	t.Helper()
	type umsClaims struct {
		Email string   `json:"email"`
		Roles []string `json:"roles"`
		jwt.RegisteredClaims
	}
	claims := umsClaims{
		Email: email, Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    "ums-test",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			ID:        "jti-test",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return signed
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestGinMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	key := generateTestKey(t)
	srv := jwksServer(t, key, "test-kid")
	defer srv.Close()

	client, err := middleware.NewUMSClient(srv.URL, time.Hour)
	require.NoError(t, err)

	tokenStr := signToken(t, key, "test-kid", "uuid-alice", "alice@x.com", []string{"admin"}, time.Hour)

	w := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(w)
	router.GET("/test", client.GinMiddleware(), func(c *gin.Context) {
		claims, ok := middleware.ClaimsFrom(c)
		require.True(t, ok)
		c.JSON(http.StatusOK, gin.H{"uuid": claims.UserUUID, "email": claims.Email})
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "uuid-alice", body["uuid"])
	assert.Equal(t, "alice@x.com", body["email"])
}

func TestGinMiddleware_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	key := generateTestKey(t)
	srv := jwksServer(t, key, "kid-1")
	defer srv.Close()

	client, err := middleware.NewUMSClient(srv.URL, time.Hour)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)
	router.GET("/test", client.GinMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGinMiddleware_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	key := generateTestKey(t)
	srv := jwksServer(t, key, "kid-1")
	defer srv.Close()

	client, err := middleware.NewUMSClient(srv.URL, time.Hour)
	require.NoError(t, err)

	// Token expired 1 hour ago
	tokenStr := signToken(t, key, "kid-1", "uuid-bob", "bob@x.com", []string{"viewer"}, -time.Hour)

	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)
	router.GET("/test", client.GinMiddleware(), func(c *gin.Context) { c.Status(http.StatusOK) })
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireRole_Granted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	key := generateTestKey(t)
	srv := jwksServer(t, key, "kid-1")
	defer srv.Close()

	client, err := middleware.NewUMSClient(srv.URL, time.Hour)
	require.NoError(t, err)

	tokenStr := signToken(t, key, "kid-1", "uuid-admin", "admin@x.com", []string{"admin"}, time.Hour)

	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)
	router.DELETE("/resource",
		client.GinMiddleware(),
		client.RequireRole("admin"),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)
	req, _ := http.NewRequest(http.MethodDelete, "/resource", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_Denied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	key := generateTestKey(t)
	srv := jwksServer(t, key, "kid-1")
	defer srv.Close()

	client, err := middleware.NewUMSClient(srv.URL, time.Hour)
	require.NoError(t, err)

	tokenStr := signToken(t, key, "kid-1", "uuid-viewer", "viewer@x.com", []string{"viewer"}, time.Hour)

	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)
	router.DELETE("/resource",
		client.GinMiddleware(),
		client.RequireRole("admin"),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)
	req, _ := http.NewRequest(http.MethodDelete, "/resource", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
