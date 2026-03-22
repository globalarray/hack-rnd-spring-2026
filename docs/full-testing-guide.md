# Full Testing Guide

## Goal

Проверить полный пользовательский сценарий платформы ПрофДНК:

1. поднятие инфраструктуры;
2. создание теста через BFF;
3. старт сессии кандидата;
4. прохождение теста;
5. генерация отчета;
6. отправка отчета на email;
7. ручной resend отчета;
8. просмотр аналитических данных.

## 1. Prerequisites

Нужны:

- `docker`
- `docker compose`
- `task`
- `openssl`
- `jq`
- `python3`

## 2. Prepare Environment

### 2.1 Generate certificates

```bash
cd /Users/globalarray/hack-rnd-2026-spring
task gen-certs OUTPUT=certs
```

Будут созданы:

- `certs/ca.crt`
- `certs/ca.key`
- `certs/server.crt`
- `certs/server.key`
- `certs/client.crt`
- `certs/client.key`

### 2.2 Create engine `.env`

Создай файл:

[`services/engine-go/.env`](/Users/globalarray/hack-rnd-2026-spring/services/engine-go/.env)

На основе примера:

[`services/engine-go/.env.example`](/Users/globalarray/hack-rnd-2026-spring/services/engine-go/.env.example)

Минимальный вариант:

```env
ENGINE_PG_HOST=postgres_engine
ENGINE_PG_PORT=5432
ENGINE_PG_USER=hack
ENGINE_PG_PASSWORD=hack
ENGINE_PG_DATABASE=enginedb

SERVICE_PORT=50036
CERTS_PATH=/etc/certs
```

Для реальной отправки отчетов добавь SMTP:

```env
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=mailer@example.com
SMTP_PASSWORD=secret
SMTP_FROM=mailer@example.com
SMTP_USE_TLS=false
DEFAULT_CLIENT_REPORT_FORMAT=client_docx
```

Если `SMTP_*` не заполнены:

- тест по прохождению завершится;
- отчет сгенерируется на уровне логики;
- но `reportDelivery.status` будет `failed`.

Важно по лимиту времени:

- `engine-go` читает канонический ключ `settings.limits.time_limit`
- для обратной совместимости также поддерживается `settings.limits.time_limit_sec`
- в примерах ниже можно оставлять `time_limit_sec`

### 2.3 Create auth-go `.env`

Создай файл:

[`services/auth-go/.env`](/Users/globalarray/hack-rnd-2026-spring/services/auth-go/.env)

На основе примера:

[`services/auth-go/.env.example`](/Users/globalarray/hack-rnd-2026-spring/services/auth-go/.env.example)

Для локального MVP достаточно:

```env
SERVICE_PORT=50037
CERTS_PATH=/etc/certs
MIGRATIONS_PATH=/app/internal/migrations/init.sql

AUTH_PG_HOST=postgres_auth
AUTH_PG_PORT=5432
AUTH_PG_USER=hack
AUTH_PG_PASSWORD=hack
AUTH_PG_DATABASE=authdb
AUTH_PG_SSL_MODE=disable

ACCESS_SECRET_KEY=change-me-access-secret
REFRESH_SECRET_KEY=change-me-refresh-secret

AUTH_BOOTSTRAP_ADMIN_EMAIL=admin@profdnk.local
AUTH_BOOTSTRAP_ADMIN_PASSWORD=admin12345
AUTH_BOOTSTRAP_ADMIN_FULL_NAME=System Administrator
AUTH_BOOTSTRAP_ADMIN_PHONE=+70000000000
AUTH_BOOTSTRAP_ADMIN_ACCESS_UNTIL=2099-12-31
AUTH_BOOTSTRAP_ADMIN_ROLE=admin
```

Что это даёт:

- при первом запуске автоматически создаётся bootstrap admin
- этот admin использует тот же endpoint логина, что и психологи
- администратор может создавать invitation link для психологов

## 3. Start the Stack

Перед стартом полезно синхронизировать сгенерированные контракты:

```bash
cd /Users/globalarray/hack-rnd-2026-spring
REBUILD_PROTO=true task proto
```

Эта команда:

- проверит наличие `protoc`;
- сгенерирует shared Go-контракты для `engine-go` и `bff-go`;
- пересоберет Python gRPC-контракт для `analytics-python`.

После этого поднимай стек:

```bash
cd /Users/globalarray/hack-rnd-2026-spring
task rebuild
```

Проверить контейнеры:

```bash
docker compose --env-file services/engine-go/.env ps
```

Ожидаемые сервисы:

- `postgres_engine`
- `postgres_auth`
- `migrations`
- `test-engine`
- `analytics-python`
- `auth-go`
- `bff-go`

Проверить BFF:

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

## 4. Authentication Flow

Если хочется прогнать весь auth-flow одной командой, можно использовать:

```bash
bash /Users/globalarray/hack-rnd-2026-spring/scripts/smoke-test-auth-bff.sh
```

### 4.1 Login As Bootstrap Admin

```bash
ADMIN_LOGIN_RESP=$(curl -s -X POST http://localhost:8080/public/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "admin@profdnk.local",
    "password": "admin12345"
  }')

echo "$ADMIN_LOGIN_RESP" | jq
ADMIN_ACCESS_TOKEN=$(echo "$ADMIN_LOGIN_RESP" | jq -r '.accessToken')
ADMIN_REFRESH_TOKEN=$(echo "$ADMIN_LOGIN_RESP" | jq -r '.refreshToken')
```

### 4.2 Create Invitation For Psychologist

```bash
INVITATION_RESP=$(curl -s -X POST http://localhost:8080/api/v1/auth/invitations \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
  -d '{
    "fullName": "Анна Смирнова",
    "phone": "+79990001122",
    "email": "anna.smirnova@example.com",
    "role": "psychologist",
    "accessUntil": "2027-12-31",
    "expiresAt": "2027-01-31T23:59:59Z"
  }')

echo "$INVITATION_RESP" | jq
INVITATION_TOKEN=$(echo "$INVITATION_RESP" | jq -r '.invitationToken')
```

Если нужно перевыпустить ссылку для того же email до регистрации, можно просто ещё раз вызвать эту же ручку: предыдущая неиспользованная invitation будет заменена новой.

### 4.3 Complete Registration As Psychologist

```bash
REGISTER_RESP=$(curl -s -X POST http://localhost:8080/public/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "token": "'"$INVITATION_TOKEN"'",
    "password": "StrongPass123"
  }')

echo "$REGISTER_RESP" | jq
PSY_ACCESS_TOKEN=$(echo "$REGISTER_RESP" | jq -r '.accessToken')
```

### 4.4 Get Current Profile

```bash
curl -s http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer $PSY_ACCESS_TOKEN" \
  | jq
```

Проверь:

- `role = "psychologist"`
- `status = "active"`
- `accessUntil` заполнен

## 5. Create Survey

Важно: поле `type` должно быть одним из строго допустимых значений:

- `single_choice`
- `multiple_choice`
- `scale`
- `text`

Значение вроде `12single_choice` не пройдет валидацию и вернет ошибку:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "invalid input: unsupported question type \"12single_choice\""
  }
}
```

```bash
ANSWER_1_UUID=$(uuidgen | tr '[:upper:]' '[:lower:]')
ANSWER_2_UUID=$(uuidgen | tr '[:upper:]' '[:lower:]')
ANSWER_3_UUID=$(uuidgen | tr '[:upper:]' '[:lower:]')
ANSWER_4_UUID=$(uuidgen | tr '[:upper:]' '[:lower:]')

CREATE_RESP=$(curl -s -X POST http://localhost:8080/api/v1/surveys \
  -H 'Content-Type: application/json' \
  -d @- <<JSON
{
  "psychologistId": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "title": "12Тест профориентации через BFF",
  "description": "12Проверка полного сценария",
  "settings": {
    "limits": {
      "time_limit_sec": 900
    }
  },
  "questions": [
    {
      "orderNum": 1,
      "type": "single_choice",
      "text": "Какой формат деятельности вам ближе?",
      "logicRules": {
        "rules": {},
        "default_next": "linear"
      },
      "answers": [
        {
          "id": "$ANSWER_1_UUID",
          "text": "Общение с людьми",
          "weight": 1,
          "categoryTag": "people"
        },
        {
          "id": "$ANSWER_2_UUID",
          "text": "Работа с данными",
          "weight": 2,
          "categoryTag": "analysis"
        }
      ]
    },
    {
      "orderNum": 2,
      "type": "single_choice",
      "text": "Что вам интереснее делать каждый день?",
      "logicRules": {
        "rules": {},
        "default_next": "linear"
      },
      "answers": [
        {
          "id": "$ANSWER_3_UUID",
          "text": "Анализировать информацию",
          "weight": 3,
          "categoryTag": "analysis"
        },
        {
          "id": "$ANSWER_4_UUID",
          "text": "Координировать процессы",
          "weight": 2,
          "categoryTag": "management"
        }
      ]
    }
  ]
}
JSON
)

echo "$CREATE_RESP" | jq
SURVEY_ID=$(echo "$CREATE_RESP" | jq -r '.surveyId')
```

Почему так:

- `answers.id` в БД глобально уникален
- если повторно запускать один и тот же `CreateSurvey` с одинаковыми `answerId`, Postgres вернет `duplicate key value violates unique constraint "answers_pkey"`
- поэтому для smoke-тестов безопаснее каждый раз генерировать новые `UUID`

Проверить, что тест появился в кабинете:

```bash
curl -s "http://localhost:8080/api/v1/surveys?psychologistId=3fa85f64-5717-4562-b3fc-2c963f66afa6" | jq
```

## 5. Start Session

```bash
START_RESP=$(curl -s -X POST http://localhost:8080/public/v1/sessions \
  -H 'Content-Type: application/json' \
  -d @- <<JSON
{
  "surveyId": "$SURVEY_ID",
  "clientMetadata": {
    "fullName": "Иван Иванов",
    "email": "ivan@example.com",
    "age": 17
  }
}
JSON
)

echo "$START_RESP" | jq
SESSION_ID=$(echo "$START_RESP" | jq -r '.sessionId')
Q1_ID=$(echo "$START_RESP" | jq -r '.firstQuestion.questionId')
A1_ID=$(echo "$START_RESP" | jq -r '.firstQuestion.answers[0].answerId')
```

Что важно проверить:

- `sessionId` заполнен;
- `firstQuestion` не пустой;
- `clientMetadata.fullName` передается на старте, потому что это имя потом должно попасть в отчет.

Если в вашей форме поле называется `fio`, это тоже поддерживается:

```json
{
  "surveyId": "YOUR_SURVEY_ID",
  "clientMetadata": {
    "fio": "Иван Иванов",
    "email": "ivan@example.com"
  }
}
```

## 6. Get Current Question

```bash
curl -s "http://localhost:8080/public/v1/sessions/$SESSION_ID/current-question" | jq
```

## 7. Submit First Answer

```bash
ANSWER1_RESP=$(curl -s -X POST "http://localhost:8080/public/v1/sessions/$SESSION_ID/answers" \
  -H 'Content-Type: application/json' \
  -d @- <<JSON
{
  "questionId": "$Q1_ID",
  "answerId": "$A1_ID"
}
JSON
)

echo "$ANSWER1_RESP" | jq
Q2_ID=$(echo "$ANSWER1_RESP" | jq -r '.nextQuestion.questionId')
A2_ID=$(echo "$ANSWER1_RESP" | jq -r '.nextQuestion.answers[0].answerId')
```

Ожидаемо:

- `isFinished=false`
- `nextQuestion` присутствует

## 8. Finish the Test

```bash
FINISH_RESP=$(curl -s -X POST "http://localhost:8080/public/v1/sessions/$SESSION_ID/answers" \
  -H 'Content-Type: application/json' \
  -d @- <<JSON
{
  "questionId": "$Q2_ID",
  "answerId": "$A2_ID"
}
JSON
)

echo "$FINISH_RESP" | jq
```

Ожидаемо:

- `isFinished=true`
- появляется `reportDelivery`

### Happy path with SMTP

Если SMTP настроен корректно:

- `reportDelivery.status = "sent"`
- `reportDelivery.email = "ivan@example.com"`

### Local path without SMTP

Если SMTP не настроен:

- `reportDelivery.status = "failed"`
- это нормально для локального smoke-test

## 9. Check Analytics

```bash
curl -s "http://localhost:8080/api/v1/sessions/$SESSION_ID/analytics" | jq
```

Ожидаемо:

- `clientMetadata.fullName` или `clientMetadata.email` присутствуют
- если на старте был передан `fullName`, `full_name` или `fio`, именно это имя попадет в сгенерированный отчет
- `responses` содержит ответы по обоим вопросам

## 10. Manual Report Resend

```bash
curl -s -X POST "http://localhost:8080/api/v1/sessions/$SESSION_ID/report/send" \
  -H 'Content-Type: application/json' \
  -d '{"reportFormat":"client_docx"}' \
  | jq
```

Если SMTP настроен:

- отчет будет повторно отправлен на email кандидата

## 11. What to Check in Logs

```bash
docker compose --env-file services/engine-go/.env logs -f bff-go
docker compose --env-file services/engine-go/.env logs -f analytics-python
docker compose --env-file services/engine-go/.env logs -f test-engine
```

Проверь:

- BFF не падает на `StartSession`
- `analytics-python` пишет `GENERATE REPORT`
- нет `INTERNAL` ошибок на `GenerateReport`
- `engine-go` возвращает сырые ответы и metadata

## 12. Contract Validation

Синтаксическая проверка `proto`:

```bash
cd /Users/globalarray/hack-rnd-2026-spring
protoc -I proto --descriptor_set_out=/tmp/auth.pb proto/auth.proto
protoc -I proto --descriptor_set_out=/tmp/test_engine.pb proto/test_engine.proto
protoc -I proto --descriptor_set_out=/tmp/analytics.pb proto/analytics.proto
```

## 13. Local Code Validation

Проверка BFF:

```bash
cd /Users/globalarray/hack-rnd-2026-spring/services/bff-go
GOCACHE=/Users/globalarray/hack-rnd-2026-spring/.gocache \
GOMODCACHE=/Users/globalarray/hack-rnd-2026-spring/.gomodcache \
GOPROXY=file:///Users/globalarray/go/pkg/mod/cache/download \
GOSUMDB=off \
go test ./...
```

Проверка Python analytics:

```bash
cd /Users/globalarray/hack-rnd-2026-spring
PYTHONPYCACHEPREFIX=/Users/globalarray/hack-rnd-2026-spring/.pycache \
python3 -m py_compile \
  services/analytics-python/analytics_server.py \
  services/analytics-python/mappers.py \
  services/analytics-python/report_generator.py
```
