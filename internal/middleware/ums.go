// Package middleware provides Gin middleware for feature_orbit_server.
// Authentication is delegated entirely to UMS: JWTs are verified locally
// using the RSA public key fetched from UMS's JWKS endpoint.
// No round-trip to UMS is required per request.
package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims holds the verified identity extracted from a UMS-issued JWT.
type Claims struct {
	UserUUID string
	Email    string
	Roles    []string
	Scopes   []string
	JTI      string
}

const ContextKeyClaims = "ums_claims"

// UMSClient verifies UMS JWTs locally using a cached JWKS public key.
type UMSClient struct {
	baseURL    string
	cacheTTL   time.Duration
	httpClient *http.Client
	mu         sync.RWMutex
	keys       map[string]*rsa.PublicKey
	fetchedAt  time.Time
}

// NewUMSClient creates a UMSClient and fetches the initial JWKS from UMS.
func NewUMSClient(baseURL string, cacheTTL time.Duration) (*UMSClient, error) {
	if cacheTTL == 0 {
		cacheTTL = time.Hour
	}
	c := &UMSClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		cacheTTL:   cacheTTL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		keys:       make(map[string]*rsa.PublicKey),
	}
	if err := c.refreshKeys(); err != nil {
		return nil, fmt.Errorf("UMSClient: initial JWKS fetch: %w", err)
	}
	return c, nil
}

// GinMiddleware returns a Gin handler that validates the Bearer JWT and injects
// Claims into the Gin context. Returns 401 if the token is missing or invalid.
func (c *UMSClient) GinMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, err := c.Verify(ctx.GetHeader("Authorization"))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    40004,
				"message": "token invalid or expired",
			})
			return
		}
		ctx.Set(ContextKeyClaims, claims)
		ctx.Set("user_uuid", claims.UserUUID)
		ctx.Set("roles", claims.Roles)
		ctx.Next()
	}
}

// RequireRole returns a Gin handler that rejects requests where the authenticated
// user does not hold at least one of the specified roles.
func (c *UMSClient) RequireRole(roles ...string) gin.HandlerFunc {
	need := make(map[string]bool, len(roles))
	for _, r := range roles {
		need[r] = true
	}
	return func(ctx *gin.Context) {
		v, exists := ctx.Get(ContextKeyClaims)
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 40005, "message": "forbidden"})
			return
		}
		for _, r := range v.(*Claims).Roles {
			if need[r] {
				ctx.Next()
				return
			}
		}
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 40005, "message": "forbidden"})
	}
}

// ClaimsFrom extracts the verified Claims from a Gin context (set by GinMiddleware).
func ClaimsFrom(ctx *gin.Context) (*Claims, bool) {
	v, ok := ctx.Get(ContextKeyClaims)
	if !ok {
		return nil, false
	}
	cl, ok := v.(*Claims)
	return cl, ok
}

// Verify validates a Bearer token string and returns the embedded Claims.
func (c *UMSClient) Verify(authHeader string) (*Claims, error) {
	tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tokenStr == "" {
		return nil, fmt.Errorf("missing token")
	}

	// Parse header to extract kid without full verification
	unverified, _, err := jwt.NewParser().ParseUnverified(tokenStr, &umsClaims{})
	if err != nil {
		return nil, fmt.Errorf("malformed token: %w", err)
	}
	kid, _ := unverified.Header["kid"].(string)

	pubKey, err := c.getKey(kid)
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenStr, &umsClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected alg: %v", t.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("token invalid: %w", err)
	}

	raw, ok := token.Claims.(*umsClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid claims")
	}
	return &Claims{
		UserUUID: raw.Subject,
		Email:    raw.Email,
		Roles:    raw.Roles,
		Scopes:   raw.Scopes,
		JTI:      raw.ID,
	}, nil
}

func (c *UMSClient) getKey(kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	key, ok := c.keys[kid]
	stale := time.Since(c.fetchedAt) > c.cacheTTL
	c.mu.RUnlock()
	if ok && !stale {
		return key, nil
	}
	// Refresh on cache miss or expiry; serve stale on failure.
	if err := c.refreshKeys(); err != nil {
		c.mu.RLock()
		key = c.keys[kid]
		c.mu.RUnlock()
		if key != nil {
			return key, nil // graceful degradation
		}
		return nil, fmt.Errorf("JWKS refresh failed and key %q not cached: %w", kid, err)
	}
	c.mu.RLock()
	key = c.keys[kid]
	c.mu.RUnlock()
	if key == nil {
		return nil, fmt.Errorf("key id %q not found in JWKS", kid)
	}
	return key, nil
}

func (c *UMSClient) refreshKeys() error {
	resp, err := c.httpClient.Get(c.baseURL + "/.well-known/jwks.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}
	newKeys := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" {
			continue
		}
		nB, _ := base64.RawURLEncoding.DecodeString(k.N)
		eB, _ := base64.RawURLEncoding.DecodeString(k.E)
		newKeys[k.Kid] = &rsa.PublicKey{
			N: new(big.Int).SetBytes(nB),
			E: int(new(big.Int).SetBytes(eB).Int64()),
		}
	}
	c.mu.Lock()
	c.keys = newKeys
	c.fetchedAt = time.Now()
	c.mu.Unlock()
	return nil
}

type umsClaims struct {
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	Scopes []string `json:"scopes,omitempty"`
	jwt.RegisteredClaims
}
