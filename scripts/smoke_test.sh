#!/usr/bin/env bash
# UMS + feature_orbit_server E2E smoke test
set -euo pipefail

UMS=${1:-http://localhost:8081}
APP=${2:-http://localhost:8080}
PASS=0; FAIL=0

check() {
  local name=$1 expected=$2; shift 2
  local actual; actual=$(curl -s -o /dev/null -w "%{http_code}" "$@")
  if [ "$actual" = "$expected" ]; then echo "  ✅ $name"; ((PASS++)); else echo "  ❌ $name (expected $expected, got $actual)"; ((FAIL++)); fi
}

echo "=== UMS ==="
check "GET /health" 200 "$UMS/health"
check "GET /jwks.json" 200 "$UMS/.well-known/jwks.json"

LOGIN=$(curl -s -X POST "$UMS/api/v1/auth/login" -H 'Content-Type: application/json' -d '{"email":"smoke@test.com","password":"Smoke@1234"}')
TOKEN=$(echo $LOGIN | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

echo "=== App ==="
check "GET /health" 200 "$APP/health"
check "GET /api/v1/features" 200 "$APP/api/v1/features"
check "GET /api/v1/me (no token)" 401 "$APP/api/v1/me"

if [ -n "$TOKEN" ]; then
  check "GET /api/v1/me (with token)" 200 -H "Authorization: Bearer $TOKEN" "$APP/api/v1/me"
fi

echo ""; echo "PASS: $PASS  FAIL: $FAIL"
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
