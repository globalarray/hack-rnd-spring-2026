# BFF API for Frontend

## Purpose

`bff-go` is the single REST entrypoint for the frontend.

It hides the internal gRPC topology and orchestrates three backend responsibilities:

1. survey management through `engine-go`
2. candidate session flow through `engine-go`
3. report generation through `analytics-python` and report delivery to candidate email through SMTP

Important: generated reports are **not stored on the BFF server disk**.  
The BFF keeps the report only in memory, immediately attaches it to an email, and sends it to `clientMetadata.email`.

## Base URL

For local development:

```text
http://localhost:8080
```

## Error Format

Every non-2xx response has a uniform body:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "surveyId must be a valid UUID"
  }
}
```

Typical codes:

- `invalid_request`
- `not_found`
- `alreadyexists`
- `failedprecondition`
- `unimplemented`
- `service_unavailable`
- `internal_error`

## Question Types

The frontend must use string values:

- `single_choice`
- `multiple_choice`
- `scale`
- `text`

## Candidate Flow

### 1. Start Session

`POST /public/v1/sessions`

Alias kept for compatibility:

`POST /public/v1/sessions/start`

Request:

```json
{
  "surveyId": "e226326e-ca21-4acc-8afa-668f4a2124f3",
  "clientMetadata": {
    "fullName": "Иван Иванов",
    "email": "ivan@example.com",
    "age": 17,
    "school": "Гимназия 12"
  }
}
```

Rules:

- `surveyId` must be a valid UUID
- `clientMetadata.email` is mandatory
- `clientMetadata.fullName`, `clientMetadata.full_name` or `clientMetadata.fio` is recommended, because this value is used in the generated report
- `clientMetadata` may contain any additional fields required by the start form

Response:

```json
{
  "sessionId": "0a98cf2b-2a23-4f95-969f-8f6ce05b51f1",
  "firstQuestion": {
    "questionId": "5b3d954f-76ad-4d40-969a-42f63b15d7dd",
    "type": "single_choice",
    "text": "Какой формат деятельности вам ближе?",
    "answers": [
      {
        "answerId": "123e4567-e89b-42d3-a456-426614174001",
        "text": "Общение с людьми"
      },
      {
        "answerId": "123e4567-e89b-42d3-a456-426614174002",
        "text": "Работа с данными"
      }
    ]
  }
}
```

### 2. Get Current Question

`GET /public/v1/sessions/{sessionId}/current-question`

Response:

```json
{
  "questionId": "5b3d954f-76ad-4d40-969a-42f63b15d7dd",
  "type": "single_choice",
  "text": "Какой формат деятельности вам ближе?",
  "answers": [
    {
      "answerId": "123e4567-e89b-42d3-a456-426614174001",
      "text": "Общение с людьми"
    }
  ]
}
```

### 3. Submit Answer

`POST /public/v1/sessions/{sessionId}/answers`

Single choice example:

```json
{
  "questionId": "5b3d954f-76ad-4d40-969a-42f63b15d7dd",
  "answerId": "123e4567-e89b-42d3-a456-426614174001"
}
```

Text question example:

```json
{
  "questionId": "6e8042cc-e0f5-4d3d-b39b-fc3d219c355f",
  "rawText": "Мне нравится разбирать сложные системы и искать закономерности"
}
```

Multiple choice example:

```json
{
  "questionId": "6e8042cc-e0f5-4d3d-b39b-fc3d219c355f",
  "answerIds": [
    "123e4567-e89b-42d3-a456-426614174010",
    "123e4567-e89b-42d3-a456-426614174011"
  ]
}
```

Request rules:

- exactly one payload must be sent: `answerId` or `rawText` or `answerIds`
- `questionId` is mandatory

Response while test is still running:

```json
{
  "nextQuestionId": "7eb44f8f-2f0f-4019-8834-45c0879ff53c",
  "isFinished": false,
  "nextQuestion": {
    "questionId": "7eb44f8f-2f0f-4019-8834-45c0879ff53c",
    "type": "single_choice",
    "text": "Что вам интереснее делать каждый день?",
    "answers": [
      {
        "answerId": "123e4567-e89b-42d3-a456-426614174101",
        "text": "Анализировать информацию"
      }
    ]
  }
}
```

Response when test is finished and report delivery succeeded:

```json
{
  "nextQuestionId": "",
  "isFinished": true,
  "reportDelivery": {
    "status": "sent",
    "email": "ivan@example.com",
    "fileName": "client-report.docx",
    "contentType": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    "errorMessage": ""
  }
}
```

Response when test is finished but report delivery failed:

```json
{
  "nextQuestionId": "",
  "isFinished": true,
  "reportDelivery": {
    "status": "failed",
    "email": "",
    "fileName": "",
    "contentType": "",
    "errorMessage": "report delivery is disabled"
  }
}
```

Important frontend behavior:

- `isFinished=true` means the session is already completed in `engine-go`
- even if email delivery fails, the answer is already accepted
- in that case the psychologist UI may call manual resend endpoint later

## Psychologist Cabinet Flow

### 1. Create Survey

`POST /api/v1/surveys`

Request:

```json
{
  "psychologistId": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "title": "Тест профориентации",
  "description": "Черновик для кабинета профориентолога",
  "settings": {
    "limits": {
      "time_limit_sec": 900
    },
    "start_form": {
      "fields": [
        "fullName",
        "email",
        "age"
      ]
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
          "id": "123e4567-e89b-42d3-a456-426614174001",
          "text": "Общение с людьми",
          "weight": 1,
          "categoryTag": "people"
        }
      ]
    }
  ]
}
```

Response:

```json
{
  "surveyId": "e226326e-ca21-4acc-8afa-668f4a2124f3"
}
```

### 2. List Surveys

`GET /api/v1/surveys?psychologistId=3fa85f64-5717-4562-b3fc-2c963f66afa6`

Response:

```json
{
  "surveys": [
    {
      "surveyId": "e226326e-ca21-4acc-8afa-668f4a2124f3",
      "title": "Тест профориентации",
      "completionsCount": 12
    }
  ]
}
```

### 3. Get Session Analytics

`GET /api/v1/sessions/{sessionId}/analytics`

Response:

```json
{
  "surveyId": "e226326e-ca21-4acc-8afa-668f4a2124f3",
  "sessionId": "0a98cf2b-2a23-4f95-969f-8f6ce05b51f1",
  "clientMetadata": {
    "fullName": "Иван Иванов",
    "email": "ivan@example.com",
    "age": 17
  },
  "responses": [
    {
      "questionId": "5b3d954f-76ad-4d40-969a-42f63b15d7dd",
      "questionType": "single_choice",
      "questionText": "Какой формат деятельности вам ближе?",
      "selectedWeight": 1,
      "categoryTag": "people",
      "rawText": ""
    }
  ]
}
```

### 4. Manual Report Resend

`POST /api/v1/sessions/{sessionId}/report/send`

Optional request body:

```json
{
  "reportFormat": "client_docx"
}
```

Allowed `reportFormat` values:

- `client_docx`
- `client_html`
- `psycho_docx`
- `psycho_html`

Response:

```json
{
  "status": "sent",
  "email": "ivan@example.com",
  "fileName": "client-report.docx",
  "contentType": "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
}
```

## Frontend Notes

1. Frontend never downloads report files directly from BFF in the main happy path.
2. The candidate receives the report by email to `clientMetadata.email`.
3. The candidate name in the report is taken from `clientMetadata.fullName`, `clientMetadata.full_name` or `clientMetadata.fio`.
4. For constructor UI, keep `settings` and `logicRules` as plain JSON objects. BFF converts them to protobuf `Struct`.
5. `SubmitAnswer` already returns the next question, so the client does not need an additional request after each answer.
6. Multiple choice payload is accepted by REST contract, but final support still depends on downstream business implementation in `engine-go`. Handle `501 Not Implemented` gracefully if such a question is encountered.

## Recommended Frontend Integration Sequence

1. Psychologist creates survey through `POST /api/v1/surveys`
2. Psychologist receives public survey link from frontend routing
3. Candidate fills start form and frontend calls `POST /public/v1/sessions`
4. Frontend renders `firstQuestion`
5. Frontend keeps calling `POST /public/v1/sessions/{sessionId}/answers`
6. While `isFinished=false`, render `nextQuestion`
7. When `isFinished=true`, show completion screen and delivery status
8. If delivery failed, psychologist may later use `POST /api/v1/sessions/{sessionId}/report/send`
