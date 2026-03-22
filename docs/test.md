# Test Requests

Готовые запросы для всех REST-ручек `bff-go`.

Если нужен полный happy-path одной командой, можно использовать готовый скрипт:

```bash
bash /Users/globalarray/hack-rnd-2026-spring/scripts/smoke-test-bff.sh
bash /Users/globalarray/hack-rnd-2026-spring/scripts/smoke-test-auth-bff.sh
```

Базовый URL:

```bash
BASE_URL=http://localhost:8080
PSYCHOLOGIST_ID=3fa85f64-5717-4562-b3fc-2c963f66afa6
```

Для auth-flow локально по умолчанию используется bootstrap admin:

```bash
ADMIN_EMAIL=admin@profdnk.local
ADMIN_PASSWORD=admin12345
```

Для повторяемого `CreateSurvey` без конфликта по `answers_pkey` сгенерируй новые `UUID`:

```bash
ANSWER_1_UUID=$(uuidgen | tr '[:upper:]' '[:lower:]')
ANSWER_2_UUID=$(uuidgen | tr '[:upper:]' '[:lower:]')
ANSWER_3_UUID=$(uuidgen | tr '[:upper:]' '[:lower:]')
ANSWER_4_UUID=$(uuidgen | tr '[:upper:]' '[:lower:]')
```

Допустимые типы вопросов:

- `single_choice`
- `multiple_choice`
- `scale`
- `text`

## 1. Health

```bash
curl -s "$BASE_URL/health" | jq
```

## 2. Login As Admin

```bash
ADMIN_LOGIN_RESP=$(curl -s -X POST "$BASE_URL/public/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "'"$ADMIN_EMAIL"'",
    "password": "'"$ADMIN_PASSWORD"'"
  }')

echo "$ADMIN_LOGIN_RESP" | jq
ADMIN_ACCESS_TOKEN=$(echo "$ADMIN_LOGIN_RESP" | jq -r '.accessToken')
ADMIN_REFRESH_TOKEN=$(echo "$ADMIN_LOGIN_RESP" | jq -r '.refreshToken')
```

## 3. Refresh Admin Token

```bash
curl -s -X POST "$BASE_URL/public/v1/auth/refresh" \
  -H 'Content-Type: application/json' \
  -d '{
    "refreshToken": "'"$ADMIN_REFRESH_TOKEN"'"
  }' \
  | jq
```

## 4. Create Psychologist Invitation

Повторный вызов для того же `email` до регистрации выдаст новую invitation и заменит предыдущую неиспользованную.

```bash
INVITATION_RESP=$(curl -s -X POST "$BASE_URL/api/v1/auth/invitations" \
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

## 5. Register Psychologist From Invitation

```bash
REGISTER_RESP=$(curl -s -X POST "$BASE_URL/public/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d '{
    "token": "'"$INVITATION_TOKEN"'",
    "password": "StrongPass123"
  }')

echo "$REGISTER_RESP" | jq
PSY_ACCESS_TOKEN=$(echo "$REGISTER_RESP" | jq -r '.accessToken')
```

## 6. Get Current User Profile

```bash
curl -s "$BASE_URL/api/v1/auth/profile" \
  -H "Authorization: Bearer $PSY_ACCESS_TOKEN" \
  | jq
```

## 7. Update Current User Profile

```bash
curl -s -X PATCH "$BASE_URL/api/v1/auth/profile" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $PSY_ACCESS_TOKEN" \
  -d '{
    "photoUrl": "https://example.com/avatar.jpg",
    "about": "Практикующий профориентолог"
  }' \
  | jq
```

## 8. Create Survey

Приватная ручка, требуется `Authorization`.

```bash
CREATE_RESP=$(curl -s -X POST "$BASE_URL/api/v1/surveys" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
  -d @- <<JSON
{
  "psychologistId": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "title": "Тест профориентации через BFF",
  "description": "Проверка полного REST-сценария",
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

## 9. List Surveys

```bash
curl -s "$BASE_URL/api/v1/surveys?psychologistId=$PSYCHOLOGIST_ID" \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
  | jq
```

## 10. Start Session

```bash
START_RESP=$(curl -s -X POST "$BASE_URL/public/v1/sessions" \
  -H 'Content-Type: application/json' \
  -d @- <<JSON
{
  "surveyId": "$SURVEY_ID",
  "clientMetadata": {
    "fullName": "Иван Иванов",
    "email": "ivan@example.com",
    "age": 17,
    "school": "Гимназия 12"
  }
}
JSON
)

echo "$START_RESP" | jq
SESSION_ID=$(echo "$START_RESP" | jq -r '.sessionId')
Q1_ID=$(echo "$START_RESP" | jq -r '.firstQuestion.questionId')
A1_ID=$(echo "$START_RESP" | jq -r '.firstQuestion.answers[0].answerId')
```

Алиас той же ручки:

```bash
curl -s -X POST "$BASE_URL/public/v1/sessions/start" \
  -H 'Content-Type: application/json' \
  -d @- <<JSON
{
  "surveyId": "$SURVEY_ID",
  "clientMetadata": {
    "fio": "Иван Иванов",
    "email": "ivan@example.com"
  }
}
JSON
| jq
```

## 11. Get Current Question

```bash
curl -s "$BASE_URL/public/v1/sessions/$SESSION_ID/current-question" | jq
```

## 12. Submit Answer: Single Choice

```bash
ANSWER1_RESP=$(curl -s -X POST "$BASE_URL/public/v1/sessions/$SESSION_ID/answers" \
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

## 13. Submit Answer: Text

Пример для текстового вопроса:

```bash
curl -s -X POST "$BASE_URL/public/v1/sessions/$SESSION_ID/answers" \
  -H 'Content-Type: application/json' \
  -d @- <<'JSON'
{
  "questionId": "6e8042cc-e0f5-4d3d-b39b-fc3d219c355f",
  "rawText": "Мне нравится анализировать сложные системы"
}
JSON
| jq
```

## 14. Submit Answer: Multiple Choice

REST-контракт это принимает, но в `engine-go` эта ветка пока может вернуть `501 Not Implemented`.

```bash
curl -s -X POST "$BASE_URL/public/v1/sessions/$SESSION_ID/answers" \
  -H 'Content-Type: application/json' \
  -d @- <<'JSON'
{
  "questionId": "6e8042cc-e0f5-4d3d-b39b-fc3d219c355f",
  "answerIds": [
    "123e4567-e89b-42d3-a456-426614174010",
    "123e4567-e89b-42d3-a456-426614174011"
  ]
}
JSON
| jq
```

## 15. Finish Survey

```bash
FINISH_RESP=$(curl -s -X POST "$BASE_URL/public/v1/sessions/$SESSION_ID/answers" \
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

## 16. Get Session Analytics

```bash
curl -s "$BASE_URL/api/v1/sessions/$SESSION_ID/analytics" \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
  | jq
```

## 17. Send Report Manually

```bash
curl -s -X POST "$BASE_URL/api/v1/sessions/$SESSION_ID/report/send" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
  -d '{"reportFormat":"client_docx"}' \
  | jq
```

Допустимые `reportFormat`:

- `client_docx`
- `client_html`
- `psycho_docx`
- `psycho_html`

## 18. Block And Unblock Psychologist

```bash
curl -s -X POST "$BASE_URL/api/v1/auth/users/block" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
  -d '{
    "email": "anna.smirnova@example.com"
  }' \
  | jq
```

```bash
curl -s -X POST "$BASE_URL/api/v1/auth/users/unblock" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
  -d '{
    "email": "anna.smirnova@example.com"
  }' \
  | jq
```

## 19. Typical Error Example

Такой запрос невалиден, потому что `type` должен быть точным строковым значением без лишних символов:

```json
{
  "type": "12single_choice"
}
```

Ожидаемый ответ:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "invalid input: unsupported question type \"12single_choice\""
  }
}
```
