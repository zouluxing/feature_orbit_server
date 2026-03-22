// Package umsclient provides a zero-dependency UMS integration SDK.
// Business services import this to verify UMS JWTs locally without UMS round-trips.
package umsclient

import (
	"context"
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
	ExpAt    time.Time
}

type Options struct {
	CacheTTL    time.Duration
	HTTPTimeout time.Duration
	ContextKey  string
}

type Client struct {
	baseURL    string
	opts       Options
	httpClient *http.Client
	mu         sync.RWMutex
	keys       map[string]*rsa.PublicKey
	fetchedAt  time.Time
}

const ContextKeyClaims = "ums_claims"

func New(baseURL string, opts Options) (*Client, error) {
	if opts.CacheTTL == 0 { opts.CacheTTL = time.Hour }
	if opts.HTTPTimeout == 0 { opts.HTTPTimeout = 10 * time.Second }
	if opts.ContextKey == "" { opts.ContextKey = ContextKeyClaims }
	c := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		opts:       opts,
		httpClient: &http.Client{Timeout: opts.HTTPTimeout},
		keys:       make(map[string]*rsa.PublicKey),
	}
	if err := c.refreshKeys(context.Background()); err != nil {
		return nil, fmt.Errorf("umsclient: initial JWKS fetch failed: %w", err)
	}
	return c, nil
}

func (c *Client) Verify(tokenStr string) (*Claims, error) {
	tokenStr = strings.TrimSpace(strings.TrimPrefix(tokenStr, "Bearer "))
	if tokenStr == "" { return nil, fmt.Errorf("empty token") }
	unverified, _, err := jwt.NewParser().ParseUnverified(tokenStr, &umsClaims{})
	if err != nil { return nil, fmt.Errorf("malformed token: %w", err) }
	kid, _ := unverified.Header["kid"].(string)
	pubKey, err := c.getKey(context.Background(), kid)
	if err != nil { return nil, err }
	token, err := jwt.ParseWithClaims(tokenStr, &umsClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil { return nil, fmt.Errorf("token invalid: %w", err) }
	raw, ok := token.Claims.(*umsClaims)
	if !ok || !token.Valid { return nil, fmt.Errorf("invalid claims") }
	return &Claims{UserUUID: raw.Subject, Email: raw.Email, Roles: raw.Roles, Scopes: raw.Scopes, JTI: raw.ID, ExpAt: raw.ExpiresAt.Time}, nil
}

func (c *Client) GinMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, err := c.Verify(ctx.GetHeader("Authorization"))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 40004, "message": "token invalid or expired"})
			return
		}
		ctx.Set(c.opts.ContextKey, claims)
		ctx.Set("user_uuid", claims.UserUUID)
		ctx.Set("roles", claims.Roles)
		ctx.Next()
	}
}

func (c *Client) RequireRole(roles ...string) gin.HandlerFunc {
	need := make(map[string]bool, len(roles))
	for _, r := range roles { need[r] = true }
	return func(ctx *gin.Context) {
		v, exists := ctx.Get(c.opts.ContextKey)
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 40005, "message": "forbidden"})
			return
		}
		for _, r := range v.(*Claims).Roles {
			if need[r] { ctx.Next(); return }
		}
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 40005, "message": "forbidden"})
	}
}

func ClaimsFrom(ctx *gin.Context) (*Claims, bool) {
	v, ok := ctx.Get(ContextKeyClaims)
	if !ok { return nil, false }
	c, ok := v.(*Claims)
	return c, ok
}

func (c *Client) getKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	key, ok := c.keys[kid]
	stale := time.Since(c.fetchedAt) > c.opts.CacheTTL
	c.mu.RUnlock()
	if ok && !stale { return key, nil }
	if err := c.refreshKeys(ctx); err != nil {
		c.mu.RLock(); key = c.keys[kid]; c.mu.RUnlock()
		if key != nil { return key, nil }
		return nil, fmt.Errorf("JWKS refresh failed and key %q not cached: %w", kid, err)
	}
	c.mu.RLock(); key = c.keys[kid]; c.mu.RUnlock()
	if key == nil { return nil, fmt.Errorf("key id %q not found in JWKS", kid) }
	return key, nil
}

func (c *Client) refreshKeys(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/.well-known/jwks.json", nil)
	resp, err := c.httpClient.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil { return err }
	newKeys := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" { continue }
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil { continue }
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil { continue }
		newKeys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: int(new(big.Int).SetBytes(eBytes).Int64())}
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
