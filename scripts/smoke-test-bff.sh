#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
PSYCHOLOGIST_ID="${PSYCHOLOGIST_ID:-3fa85f64-5717-4562-b3fc-2c963f66afa6}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@profdnk.local}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin12345}"
CLIENT_NAME="${CLIENT_NAME:-Иван Иванов}"
CLIENT_EMAIL="${CLIENT_EMAIL:-ivan@example.com}"
REPORT_FORMAT="${REPORT_FORMAT:-client_docx}"

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
    response="$(curl "${curl_args[@]}")"
  else
    response="$(curl "${curl_args[@]}")"
  fi

  status="${response##*$'\n'}"
  RESPONSE_BODY="${response%$'\n'*}"
  RESPONSE_STATUS="$status"

  if [[ "$status" -lt 200 || "$status" -ge 300 ]]; then
    echo "request failed: $method $url -> HTTP $status" >&2
    pretty_print "$RESPONSE_BODY" >&2
    exit 1
  fi
}

request_json_soft() {
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
    response="$(curl "${curl_args[@]}")"
  else
    response="$(curl "${curl_args[@]}")"
  fi

  status="${response##*$'\n'}"
  RESPONSE_BODY="${response%$'\n'*}"
  RESPONSE_STATUS="$status"
}

require_cmd curl
require_cmd jq
require_cmd uuidgen

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
ADMIN_AUTHORIZATION="Bearer $ADMIN_ACCESS_TOKEN"

ANSWER_1_UUID="$(uuidgen | tr '[:upper:]' '[:lower:]')"
ANSWER_2_UUID="$(uuidgen | tr '[:upper:]' '[:lower:]')"
ANSWER_3_UUID="$(uuidgen | tr '[:upper:]' '[:lower:]')"
ANSWER_4_UUID="$(uuidgen | tr '[:upper:]' '[:lower:]')"

CREATE_PAYLOAD="$(jq -n \
  --arg psychologistId "$PSYCHOLOGIST_ID" \
  --arg answer1 "$ANSWER_1_UUID" \
  --arg answer2 "$ANSWER_2_UUID" \
  --arg answer3 "$ANSWER_3_UUID" \
  --arg answer4 "$ANSWER_4_UUID" \
  '{
    psychologistId: $psychologistId,
    title: "Smoke test survey via BFF",
    description: "Auto-generated smoke test for the BFF and engine stack",
    settings: {
      limits: {
        time_limit: 900
      }
    },
    questions: [
      {
        orderNum: 1,
        type: "single_choice",
        text: "Какой формат деятельности вам ближе?",
        logicRules: {
          rules: {},
          default_next: "linear"
        },
        answers: [
          {
            id: $answer1,
            text: "Общение с людьми",
            weight: 1,
            categoryTag: "people"
          },
          {
            id: $answer2,
            text: "Работа с данными",
            weight: 2,
            categoryTag: "analysis"
          }
        ]
      },
      {
        orderNum: 2,
        type: "single_choice",
        text: "Что вам интереснее делать каждый день?",
        logicRules: {
          rules: {},
          default_next: "linear"
        },
        answers: [
          {
            id: $answer3,
            text: "Анализировать информацию",
            weight: 3,
            categoryTag: "analysis"
          },
          {
            id: $answer4,
            text: "Координировать процессы",
            weight: 2,
            categoryTag: "management"
          }
        ]
      }
    ]
  }')"

print_step "Create Survey"
request_json POST "$BASE_URL/api/v1/surveys" "$CREATE_PAYLOAD" "$ADMIN_AUTHORIZATION"
pretty_print "$RESPONSE_BODY"
SURVEY_ID="$(printf '%s' "$RESPONSE_BODY" | jq -r '.surveyId')"

print_step "List Surveys"
request_json GET "$BASE_URL/api/v1/surveys?psychologistId=$PSYCHOLOGIST_ID" "" "$ADMIN_AUTHORIZATION"
pretty_print "$RESPONSE_BODY"

START_PAYLOAD="$(jq -n \
  --arg surveyId "$SURVEY_ID" \
  --arg fullName "$CLIENT_NAME" \
  --arg email "$CLIENT_EMAIL" \
  '{
    surveyId: $surveyId,
    clientMetadata: {
      fullName: $fullName,
      email: $email,
      source: "smoke-test-script"
    }
  }')"

print_step "Start Session"
request_json POST "$BASE_URL/public/v1/sessions" "$START_PAYLOAD"
pretty_print "$RESPONSE_BODY"
SESSION_ID="$(printf '%s' "$RESPONSE_BODY" | jq -r '.sessionId')"
Q1_ID="$(printf '%s' "$RESPONSE_BODY" | jq -r '.firstQuestion.questionId')"
A1_ID="$(printf '%s' "$RESPONSE_BODY" | jq -r '.firstQuestion.answers[0].answerId')"

print_step "Get Current Question"
request_json GET "$BASE_URL/public/v1/sessions/$SESSION_ID/current-question"
pretty_print "$RESPONSE_BODY"

ANSWER1_PAYLOAD="$(jq -n \
  --arg questionId "$Q1_ID" \
  --arg answerId "$A1_ID" \
  '{questionId: $questionId, answerId: $answerId}')"

print_step "Submit First Answer"
request_json POST "$BASE_URL/public/v1/sessions/$SESSION_ID/answers" "$ANSWER1_PAYLOAD"
pretty_print "$RESPONSE_BODY"
Q2_ID="$(printf '%s' "$RESPONSE_BODY" | jq -r '.nextQuestion.questionId')"
A2_ID="$(printf '%s' "$RESPONSE_BODY" | jq -r '.nextQuestion.answers[0].answerId')"

FINISH_PAYLOAD="$(jq -n \
  --arg questionId "$Q2_ID" \
  --arg answerId "$A2_ID" \
  '{questionId: $questionId, answerId: $answerId}')"

print_step "Finish Session"
request_json POST "$BASE_URL/public/v1/sessions/$SESSION_ID/answers" "$FINISH_PAYLOAD"
pretty_print "$RESPONSE_BODY"

print_step "Get Session Analytics"
request_json GET "$BASE_URL/api/v1/sessions/$SESSION_ID/analytics" "" "$ADMIN_AUTHORIZATION"
pretty_print "$RESPONSE_BODY"

REPORT_PAYLOAD="$(jq -n --arg reportFormat "$REPORT_FORMAT" '{reportFormat: $reportFormat}')"

print_step "Manual Report Send"
request_json_soft POST "$BASE_URL/api/v1/sessions/$SESSION_ID/report/send" "$REPORT_PAYLOAD" "$ADMIN_AUTHORIZATION"
pretty_print "$RESPONSE_BODY"
if [[ "$RESPONSE_STATUS" -lt 200 || "$RESPONSE_STATUS" -ge 300 ]]; then
  echo "report resend returned HTTP $RESPONSE_STATUS; this is expected when SMTP is not configured" >&2
fi

print_step "Summary"
cat <<EOF
SURVEY_ID=$SURVEY_ID
SESSION_ID=$SESSION_ID
Q1_ID=$Q1_ID
Q2_ID=$Q2_ID
EOF
