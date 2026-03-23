// Package middleware provides UMS JWT verification middleware.
// ADR-005: feature_orbit_server delegates all auth to UMS via JWKS local verification.
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

type Claims struct {
	UserUUID string
	Email    string
	Roles    []string
	Scopes   []string
	JTI      string
}

const ContextKeyClaims = "ums_claims"

type UMSClient struct {
	baseURL    string
	cacheTTL   time.Duration
	httpClient *http.Client
	mu         sync.RWMutex
	keys       map[string]*rsa.PublicKey
	fetchedAt  time.Time
}

func NewUMSClient(baseURL string, cacheTTL time.Duration) (*UMSClient, error) {
	if cacheTTL == 0 { cacheTTL = time.Hour }
	c := &UMSClient{
		baseURL: strings.TrimRight(baseURL, "/"), cacheTTL: cacheTTL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		keys: make(map[string]*rsa.PublicKey),
	}
	if err := c.refreshKeys(); err != nil { return nil, fmt.Errorf("UMSClient: initial JWKS fetch: %w", err) }
	return c, nil
}

func (c *UMSClient) GinMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, err := c.Verify(ctx.GetHeader("Authorization"))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 40004, "message": "token invalid or expired"})
			return
		}
		ctx.Set(ContextKeyClaims, claims)
		ctx.Set("user_uuid", claims.UserUUID)
		ctx.Set("roles", claims.Roles)
		ctx.Next()
	}
}

func (c *UMSClient) RequireRole(roles ...string) gin.HandlerFunc {
	need := make(map[string]bool, len(roles))
	for _, r := range roles { need[r] = true }
	return func(ctx *gin.Context) {
		v, exists := ctx.Get(ContextKeyClaims)
		if !exists { ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 40005, "message": "forbidden"}); return }
		for _, r := range v.(*Claims).Roles {
			if need[r] { ctx.Next(); return }
		}
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 40005, "message": "forbidden"})
	}
}

func ClaimsFrom(ctx *gin.Context) (*Claims, bool) {
	v, ok := ctx.Get(ContextKeyClaims)
	if !ok { return nil, false }
	cl, ok := v.(*Claims)
	return cl, ok
}

func (c *UMSClient) Verify(authHeader string) (*Claims, error) {
	tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tokenStr == "" { return nil, fmt.Errorf("missing token") }
	unverified, _, err := jwt.NewParser().ParseUnverified(tokenStr, &umsClaims{})
	if err != nil { return nil, fmt.Errorf("malformed token: %w", err) }
	kid, _ := unverified.Header["kid"].(string)
	pubKey, err := c.getKey(kid)
	if err != nil { return nil, err }
	token, err := jwt.ParseWithClaims(tokenStr, &umsClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok { return nil, fmt.Errorf("unexpected alg: %v", t.Header["alg"]) }
		return pubKey, nil
	})
	if err != nil { return nil, fmt.Errorf("token invalid: %w", err) }
	raw, ok := token.Claims.(*umsClaims)
	if !ok || !token.Valid { return nil, fmt.Errorf("invalid claims") }
	return &Claims{UserUUID: raw.Subject, Email: raw.Email, Roles: raw.Roles, Scopes: raw.Scopes, JTI: raw.ID}, nil
}

func (c *UMSClient) getKey(kid string) (*rsa.PublicKey, error) {
	c.mu.RLock(); key, ok := c.keys[kid]; stale := time.Since(c.fetchedAt) > c.cacheTTL; c.mu.RUnlock()
	if ok && !stale { return key, nil }
	if err := c.refreshKeys(); err != nil {
		c.mu.RLock(); key = c.keys[kid]; c.mu.RUnlock()
		if key != nil { return key, nil }
		return nil, fmt.Errorf("JWKS refresh failed and key %q not cached: %w", kid, err)
	}
	c.mu.RLock(); key = c.keys[kid]; c.mu.RUnlock()
	if key == nil { return nil, fmt.Errorf("key id %q not found in JWKS", kid) }
	return key, nil
}

func (c *UMSClient) refreshKeys() error {
	resp, err := c.httpClient.Get(c.baseURL + "/.well-known/jwks.json")
	if err != nil { return err }
	defer resp.Body.Close()
	var jwks struct { Keys []struct { Kid, Kty, N, E string `json:"kid,kty,n,e"` } `json:"keys"` }
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil { return err }
	newKeys := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" { continue }
		nB, _ := base64.RawURLEncoding.DecodeString(k.N)
		eB, _ := base64.RawURLEncoding.DecodeString(k.E)
		newKeys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(nB), E: int(new(big.Int).SetBytes(eB).Int64())}
	}
	c.mu.Lock(); c.keys = newKeys; c.fetchedAt = time.Now(); c.mu.Unlock()
	return nil
}

type umsClaims struct {
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	Scopes []string `json:"scopes,omitempty"`
	jwt.RegisteredClaims
}
