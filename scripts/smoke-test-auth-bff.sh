#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@profdnk.local}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin12345}"
PSYCHOLOGIST_EMAIL="${PSYCHOLOGIST_EMAIL:-}"
PSYCHOLOGIST_PASSWORD="${PSYCHOLOGIST_PASSWORD:-psychologist123}"
PSYCHOLOGIST_FULL_NAME="${PSYCHOLOGIST_FULL_NAME:-Анна Психолог}"
PSYCHOLOGIST_PHONE="${PSYCHOLOGIST_PHONE:-+79990001122}"
PSYCHOLOGIST_ABOUT="${PSYCHOLOGIST_ABOUT:-Психолог профориентационных тестов}"

require_cmd() {
  local name="$1"
  if ! command -v "$name" >/dev/null 2>&1; then
    echo "missing required command: $name" >&2
    exit 1
  fi
}

print_step() {
  printf '\n==> %s\n' "$1"
}

pretty_print() {
  local payload="$1"
  if command -v jq >/dev/null 2>&1; then
    printf '%s\n' "$payload" | jq .
  else
    printf '%s\n' "$payload"
  fi
}

request_json() {
  local method="$1"
  local url="$2"
  local body="${3:-}"
  local authorization="${4:-}"
  local response
  local status
  local curl_args=(
    -sS
    -w $'\n%{http_code}'
    -X "$method"
    "$url"
    -H 'Content-Type: application/json'
  )

  if [[ -n "$authorization" ]]; then
    curl_args+=(-H "Authorization: $authorization")
  fi

  if [[ -n "$body" ]]; then
    curl_args+=(--data-binary "$body")
  fi

  response="$(curl "${curl_args[@]}")"
  status="${response##*$'\n'}"
  RESPONSE_BODY="${response%$'\n'*}"
  RESPONSE_STATUS="$status"

  if [[ "$status" -lt 200 || "$status" -ge 300 ]]; then
    echo "request failed: $method $url -> HTTP $status" >&2
    pretty_print "$RESPONSE_BODY" >&2
    exit 1
  fi
}

future_dates() {
  python3 - <<'PY'
from datetime import datetime, timedelta, timezone

now = datetime.now(timezone.utc)
access_until = (now + timedelta(days=30)).strftime("%Y-%m-%d")
expires_at = (now + timedelta(days=7)).replace(microsecond=0).isoformat().replace("+00:00", "Z")

print(access_until)
print(expires_at)
PY
}

require_cmd curl
require_cmd jq
require_cmd uuidgen
require_cmd python3

if [[ -z "$PSYCHOLOGIST_EMAIL" ]]; then
  PSYCHOLOGIST_EMAIL="psychologist.$(uuidgen | tr '[:upper:]' '[:lower:]')@example.com"
fi

readarray -t FUTURE_DATES < <(future_dates)
ACCESS_UNTIL="${FUTURE_DATES[0]}"
EXPIRES_AT="${FUTURE_DATES[1]}"

print_step "Health"
request_json GET "$BASE_URL/health"
pretty_print "$RESPONSE_BODY"

ADMIN_LOGIN_PAYLOAD="$(jq -n \
  --arg email "$ADMIN_EMAIL" \
  --arg password "$ADMIN_PASSWORD" \
  '{email: $email, password: $password}')"

print_step "Admin Login"
request_json POST "$BASE_URL/public/v1/auth/login" "$ADMIN_LOGIN_PAYLOAD"
pretty_print "$RESPONSE_BODY"
ADMIN_ACCESS_TOKEN="$(printf '%s' "$RESPONSE_BODY" | jq -r '.accessToken')"
ADMIN_REFRESH_TOKEN="$(printf '%s' "$RESPONSE_BODY" | jq -r '.refreshToken')"
ADMIN_AUTHORIZATION="Bearer $ADMIN_ACCESS_TOKEN"

INVITATION_PAYLOAD="$(jq -n \
  --arg fullName "$PSYCHOLOGIST_FULL_NAME" \
  --arg phone "$PSYCHOLOGIST_PHONE" \
  --arg email "$PSYCHOLOGIST_EMAIL" \
  --arg role "psychologist" \
  --arg accessUntil "$ACCESS_UNTIL" \
  --arg expiresAt "$EXPIRES_AT" \
  '{
    fullName: $fullName,
    phone: $phone,
    email: $email,
    role: $role,
    accessUntil: $accessUntil,
    expiresAt: $expiresAt
  }')"

print_step "Create Invitation"
request_json POST "$BASE_URL/api/v1/auth/invitations" "$INVITATION_PAYLOAD" "$ADMIN_AUTHORIZATION"
pretty_print "$RESPONSE_BODY"
INVITATION_TOKEN="$(printf '%s' "$RESPONSE_BODY" | jq -r '.invitationToken')"
INVITATION_URL="$(printf '%s' "$RESPONSE_BODY" | jq -r '.invitationUrl')"

REGISTER_PAYLOAD="$(jq -n \
  --arg token "$INVITATION_TOKEN" \
  --arg password "$PSYCHOLOGIST_PASSWORD" \
  '{token: $token, password: $password}')"

print_step "Register Psychologist"
request_json POST "$BASE_URL/public/v1/auth/register" "$REGISTER_PAYLOAD"
pretty_print "$RESPONSE_BODY"
PSYCHOLOGIST_ACCESS_TOKEN="$(printf '%s' "$RESPONSE_BODY" | jq -r '.accessToken')"
PSYCHOLOGIST_REFRESH_TOKEN="$(printf '%s' "$RESPONSE_BODY" | jq -r '.refreshToken')"
PSYCHOLOGIST_AUTHORIZATION="Bearer $PSYCHOLOGIST_ACCESS_TOKEN"

print_step "Refresh Admin Token"
request_json POST "$BASE_URL/public/v1/auth/refresh" "$(jq -n --arg refreshToken "$ADMIN_REFRESH_TOKEN" '{refreshToken: $refreshToken}')"
pretty_print "$RESPONSE_BODY"

print_step "Get Admin Profile"
request_json GET "$BASE_URL/api/v1/auth/profile" "" "$ADMIN_AUTHORIZATION"
pretty_print "$RESPONSE_BODY"

print_step "Get Psychologist Profile"
request_json GET "$BASE_URL/api/v1/auth/profile" "" "$PSYCHOLOGIST_AUTHORIZATION"
pretty_print "$RESPONSE_BODY"
PSYCHOLOGIST_USER_ID="$(printf '%s' "$RESPONSE_BODY" | jq -r '.id')"

UPDATE_PROFILE_PAYLOAD="$(jq -n \
  --arg about "$PSYCHOLOGIST_ABOUT" \
  '{about: $about, photoUrl: ""}')"

print_step "Update Psychologist Profile"
request_json PATCH "$BASE_URL/api/v1/auth/profile" "$UPDATE_PROFILE_PAYLOAD" "$PSYCHOLOGIST_AUTHORIZATION"
pretty_print "$RESPONSE_BODY"

print_step "Get Public Psychologist Profile"
request_json GET "$BASE_URL/public/v1/profiles/$PSYCHOLOGIST_USER_ID"
pretty_print "$RESPONSE_BODY"

print_step "Block Psychologist"
request_json POST "$BASE_URL/api/v1/auth/users/block" "$(jq -n --arg email "$PSYCHOLOGIST_EMAIL" '{email: $email}')" "$ADMIN_AUTHORIZATION"
pretty_print "$RESPONSE_BODY"

print_step "Unblock Psychologist"
request_json POST "$BASE_URL/api/v1/auth/users/unblock" "$(jq -n --arg email "$PSYCHOLOGIST_EMAIL" '{email: $email}')" "$ADMIN_AUTHORIZATION"
pretty_print "$RESPONSE_BODY"

print_step "Refresh Psychologist Token"
request_json POST "$BASE_URL/public/v1/auth/refresh" "$(jq -n --arg refreshToken "$PSYCHOLOGIST_REFRESH_TOKEN" '{refreshToken: $refreshToken}')"
pretty_print "$RESPONSE_BODY"

print_step "Summary"
cat <<EOF
ADMIN_EMAIL=$ADMIN_EMAIL
PSYCHOLOGIST_EMAIL=$PSYCHOLOGIST_EMAIL
PSYCHOLOGIST_USER_ID=$PSYCHOLOGIST_USER_ID
INVITATION_TOKEN=$INVITATION_TOKEN
INVITATION_URL=$INVITATION_URL
EOF
