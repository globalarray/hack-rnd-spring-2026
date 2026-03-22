# Test Requests

Готовые запросы для всех REST-ручек `bff-go`.

Базовый URL:

```bash
BASE_URL=http://localhost:8080
PSYCHOLOGIST_ID=3fa85f64-5717-4562-b3fc-2c963f66afa6
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

## 2. Create Survey

```bash
CREATE_RESP=$(curl -s -X POST "$BASE_URL/api/v1/surveys" \
  -H 'Content-Type: application/json' \
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
          "id": "'"$ANSWER_1_UUID"'",
          "text": "Общение с людьми",
          "weight": 1,
          "categoryTag": "people"
        },
        {
          "id": "'"$ANSWER_2_UUID"'",
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
          "id": "'"$ANSWER_3_UUID"'",
          "text": "Анализировать информацию",
          "weight": 3,
          "categoryTag": "analysis"
        },
        {
          "id": "'"$ANSWER_4_UUID"'",
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

## 3. List Surveys

```bash
curl -s "$BASE_URL/api/v1/surveys?psychologistId=$PSYCHOLOGIST_ID" | jq
```

## 4. Start Session

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

## 5. Get Current Question

```bash
curl -s "$BASE_URL/public/v1/sessions/$SESSION_ID/current-question" | jq
```

## 6. Submit Answer: Single Choice

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

## 7. Submit Answer: Text

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

## 8. Submit Answer: Multiple Choice

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

## 9. Finish Survey

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

## 10. Get Session Analytics

```bash
curl -s "$BASE_URL/api/v1/sessions/$SESSION_ID/analytics" | jq
```

## 11. Send Report Manually

```bash
curl -s -X POST "$BASE_URL/api/v1/sessions/$SESSION_ID/report/send" \
  -H 'Content-Type: application/json' \
  -d '{"reportFormat":"client_docx"}' \
  | jq
```

Допустимые `reportFormat`:

- `client_docx`
- `client_html`
- `psycho_docx`
- `psycho_html`

## 12. Typical Error Example

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
