package service_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zouluxing/ums/pkg/utils"
)

func newTestManager(t *testing.T) *utils.JWTManager {
	t.Helper()
	dir := t.TempDir()
	priv, pub := dir+"/private.pem", dir+"/public.pem"
	require.NoError(t, utils.GenerateRSAKeyPairFiles(priv, pub))
	mgr, err := utils.NewJWTManager(priv, pub, "ums-test", "kid-1", 15*time.Minute, 7*24*time.Hour)
	require.NoError(t, err)
	return mgr
}

func TestIssueAndVerify_RoundTrip(t *testing.T) {
	mgr := newTestManager(t)
	token, jti, err := mgr.IssueAccessToken("uuid-123", "alice@x.com", []string{"admin"}, []string{"profile"})
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, jti)
	claims, err := mgr.VerifyAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, "uuid-123", claims.Subject)
	assert.Equal(t, "alice@x.com", claims.Email)
	assert.Equal(t, []string{"admin"}, claims.Roles)
	assert.Equal(t, jti, claims.ID)
}

func TestVerify_InvalidSignature(t *testing.T) {
	mgr := newTestManager(t)
	token, _, _ := mgr.IssueAccessToken("u1", "u@x.com", nil, nil)
	tampered := token[:len(token)-1] + "X"
	_, err := mgr.VerifyAccessToken(tampered)
	assert.Error(t, err)
}

func TestVerify_Expired(t *testing.T) {
	dir := t.TempDir()
	priv, pub := dir+"/priv.pem", dir+"/pub.pem"
	require.NoError(t, utils.GenerateRSAKeyPairFiles(priv, pub))
	mgr, err := utils.NewJWTManager(priv, pub, "ums", "k1", time.Nanosecond, time.Minute)
	require.NoError(t, err)
	token, _, _ := mgr.IssueAccessToken("u1", "u@x.com", nil, nil)
	time.Sleep(2 * time.Millisecond)
	_, err = mgr.VerifyAccessToken(token)
	assert.Error(t, err)
}

func TestPublicKeyComponents_NonEmpty(t *testing.T) {
	mgr := newTestManager(t)
	assert.NotEmpty(t, mgr.PublicKeyN())
	assert.NotEmpty(t, mgr.PublicKeyE())
	assert.Equal(t, "kid-1", mgr.KeyID())
	assert.Equal(t, 15*time.Minute, mgr.AccessTTL())
}

func TestSHA256Base64_Deterministic(t *testing.T) {
	assert.Equal(t, utils.SHA256Base64("hello"), utils.SHA256Base64("hello"))
	assert.NotEqual(t, utils.SHA256Base64("a"), utils.SHA256Base64("b"))
}

func TestDefaultInt(t *testing.T) {
	assert.Equal(t, 20, utils.DefaultInt(0, 20))
	assert.Equal(t, 5, utils.DefaultInt(5, 20))
}

func TestOffset(t *testing.T) {
	assert.Equal(t, 0, utils.Offset(1, 20))
	assert.Equal(t, 20, utils.Offset(2, 20))
}

func TestGenerateRSAKeyPairFiles(t *testing.T) {
	dir := t.TempDir()
	priv, pub := dir+"/p.pem", dir+"/pub.pem"
	require.NoError(t, utils.GenerateRSAKeyPairFiles(priv, pub))
	_, err := utils.NewJWTManager(priv, pub, "iss", "kid", time.Minute, time.Hour)
	require.NoError(t, err)
}
