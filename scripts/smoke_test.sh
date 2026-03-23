#!/usr/bin/env bash
# UMS + feature_orbit_server E2E smoke test
# Usage: ./scripts/smoke_test.sh [UMS_URL] [APP_URL]
set -euo pipefail

UMS=${1:-http://localhost:8081}
APP=${2:-http://localhost:8080}
PASS=0
FAIL=0

check() {
  local name=$1 expected=$2
  shift 2
  local actual
  actual=$(curl -s -o /dev/null -w "%{http_code}" "$@")
  if [ "$actual" = "$expected" ]; then
    echo "  ✅ $name ($actual)"
    ((PASS++))
  else
    echo "  ❌ $name — expected $expected, got $actual"
    ((FAIL++))
  fi
}

echo "=== UMS Health ==="
check "GET /health" 200 "$UMS/health"
check "GET /.well-known/jwks.json" 200 "$UMS/.well-known/jwks.json"

echo "=== Auth ==="
REGISTER=$(curl -s -X POST "$UMS/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d '{"email":"smoke@test.com","username":"smoketest","password":"Smoke@1234"}')
echo "  Register: $(echo $REGISTER | grep -o '"code":[0-9]*')"

LOGIN=$(curl -s -X POST "$UMS/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"smoke@test.com","password":"Smoke@1234"}')
TOKEN=$(echo $LOGIN | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "  ❌ Login failed — no token returned"
  ((FAIL++))
else
  echo "  ✅ Login (token obtained)"
  ((PASS++))
fi

echo "=== feature_orbit_server ==="
check "GET /health" 200 "$APP/health"
check "GET /api/v1/features (public)" 200 "$APP/api/v1/features"
check "GET /api/v1/me (no token)" 401 "$APP/api/v1/me"

if [ -n "$TOKEN" ]; then
  check "GET /api/v1/me (with token)" 200 \
    -H "Authorization: Bearer $TOKEN" "$APP/api/v1/me"
  check "POST /api/v1/features (with token)" 201 \
    -X POST -H "Authorization: Bearer $TOKEN" \
    -H 'Content-Type: application/json' \
    -d '{"name":"Smoke Feature"}' "$APP/api/v1/features"
fi

echo ""
echo "=== Results ==="
echo "  PASS: $PASS"
echo "  FAIL: $FAIL"
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
