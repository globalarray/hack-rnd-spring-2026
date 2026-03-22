# System Architecture

## Purpose

Этот документ описывает полную схему работы платформы ПрофДНК в текущем MVP-состоянии.

Система построена как микросервисная архитектура без брокера сообщений.

Основной принцип:

- фронт общается только с `bff-go` по REST;
- внутренние Go-сервисы общаются по `gRPC + mTLS`;
- генерация отчёта выполняется отдельным Python-сервисом;
- отчёт не сохраняется на диск сервера, а отправляется клиенту по email.

## High-Level Scheme

```mermaid
flowchart LR
    Admin["Администратор"] --> FE["Frontend"]
    Psychologist["Психолог"] --> FE
    Candidate["Кандидат"] --> FE

    FE -->|REST/JSON| BFF["bff-go<br/>единая точка входа"]

    BFF -->|gRPC + mTLS| AUTH["auth-go<br/>авторизация и lifecycle аккаунтов"]
    BFF -->|gRPC + mTLS| ENGINE["engine-go<br/>конструктор тестов и прохождение"]
    BFF -->|gRPC| ANALYTICS["analytics-python<br/>генерация отчётов"]
    BFF -->|SMTP| MAIL["Почтовый сервер"]

    AUTH -->|SQL| AUTHDB[("postgres_auth")]
    ENGINE -->|SQL| ENGINEDB[("postgres_engine")]
    MIGRATIONS["migrations"] -->|SQL migrations| ENGINEDB
```

## Runtime Boundaries

### External Entry Point

Единственная публичная backend-точка входа:

- `bff-go`

Он отвечает за:

- REST API для frontend;
- orchestration между микросервисами;
- валидацию и нормализацию запросов;
- преобразование внутренних gRPC-ошибок в frontend-friendly HTTP;
- отправку готового отчёта на email.

### Internal Services

#### `auth-go`

Отвечает за:

- единый вход администратора и психолога;
- access/refresh token;
- invitation-based onboarding психологов;
- bootstrap admin при первом запуске;
- блокировку, разблокировку и автоматическую деактивацию аккаунтов по `access_until`.

#### `engine-go`

Отвечает за:

- создание тестов;
- хранение вопросов и ответов;
- запуск сессий;
- возврат текущего вопроса;
- приём ответов;
- сбор аналитических данных по прохождению.

#### `analytics-python`

Отвечает за:

- сборку клиентского или психологического отчёта из сырых ответов;
- генерацию бинарного файла отчёта для последующей email-доставки.

## Main User Flows

### 1. Admin Onboarding Flow

```mermaid
sequenceDiagram
    participant Boot as Первый запуск системы
    participant Auth as auth-go
    participant AuthDB as postgres_auth

    Boot->>Auth: start service
    Auth->>AuthDB: run schema init
    Auth->>AuthDB: check bootstrap admin by email
    alt admin does not exist
        Auth->>AuthDB: create bootstrap admin from env
    else admin exists
        Auth-->>Boot: reuse existing admin
    end
```

### 2. Admin Creates Psychologist Invitation

```mermaid
sequenceDiagram
    participant Admin as Администратор
    participant FE as Frontend
    participant BFF as bff-go
    participant Auth as auth-go
    participant AuthDB as postgres_auth

    Admin->>FE: login form
    FE->>BFF: POST /public/v1/auth/login
    BFF->>Auth: Login
    Auth->>AuthDB: validate credentials
    Auth-->>BFF: access + refresh token
    BFF-->>FE: tokens + role=admin

    Admin->>FE: create invitation
    FE->>BFF: POST /api/v1/auth/invitations
    BFF->>Auth: CreateInvitation
    Auth->>AuthDB: store invitation
    Auth-->>BFF: invitation token
    BFF-->>FE: invitationUrl = /invitations/{uuid}
```

### 3. Psychologist Completes Registration

```mermaid
sequenceDiagram
    participant Psych as Психолог
    participant FE as Frontend
    participant BFF as bff-go
    participant Auth as auth-go
    participant AuthDB as postgres_auth

    Psych->>FE: open invitation URL
    Psych->>FE: enter password
    FE->>BFF: POST /public/v1/auth/register
    BFF->>Auth: Register
    Auth->>AuthDB: validate invitation
    Auth->>AuthDB: create psychologist account
    Auth->>AuthDB: mark invitation as used
    Auth-->>BFF: access + refresh token
    BFF-->>FE: authenticated psychologist session
```

### 4. Psychologist Creates Survey

```mermaid
sequenceDiagram
    participant Psych as Психолог
    participant FE as Frontend Constructor
    participant BFF as bff-go
    participant Engine as engine-go
    participant EngineDB as postgres_engine

    Psych->>FE: drag and drop constructor
    FE->>BFF: POST /api/v1/surveys
    BFF->>Engine: CreateSurvey
    Engine->>EngineDB: save survey, questions, answers
    Engine-->>BFF: survey_id
    BFF-->>FE: created survey
```

### 5. Candidate Passes Test and Receives Report

```mermaid
sequenceDiagram
    participant Candidate as Кандидат
    participant FE as Candidate Frontend
    participant BFF as bff-go
    participant Engine as engine-go
    participant EngineDB as postgres_engine
    participant Analytics as analytics-python
    participant Mail as SMTP

    Candidate->>FE: open public survey link
    Candidate->>FE: fill start form with email and FIO
    FE->>BFF: POST /public/v1/sessions
    BFF->>Engine: StartSession
    Engine->>EngineDB: create session
    Engine-->>BFF: session_id + first_question
    BFF-->>FE: first question

    loop until finished
        FE->>BFF: POST /public/v1/sessions/{id}/answers
        BFF->>Engine: SubmitAnswer
        Engine->>EngineDB: save response
        Engine-->>BFF: next_question or finished
        BFF-->>FE: next step
    end

    BFF->>Engine: GetSessionDataForAnalytics
    Engine->>EngineDB: load session analytics
    Engine-->>BFF: client metadata + responses
    BFF->>Analytics: GenerateReport
    Analytics-->>BFF: binary report
    BFF->>Mail: send email with attachment
    BFF-->>FE: reportDelivery status
```

## Data Ownership

### `postgres_auth`

Хранит:

- пользователей;
- роли;
- статусы аккаунтов;
- refresh token;
- invitation records для психологов;
- сроки действия учётных записей.

### `postgres_engine`

Хранит:

- тесты;
- вопросы;
- ответы;
- сессии прохождения;
- client metadata;
- responses;
- агрегированные данные для аналитики.

### `analytics-python`

Не является системным источником правды.

Он:

- не хранит доменные данные;
- получает сырой analytics payload от `bff-go`;
- возвращает готовый отчёт в память.

## Transport and Security Matrix

| Segment | Protocol | Security | Status |
|---|---|---|---|
| Frontend -> BFF | REST/JSON | публичный API, auth token | implemented |
| BFF -> auth-go | gRPC | mTLS | implemented |
| BFF -> engine-go | gRPC | mTLS | implemented |
| BFF -> analytics-python | gRPC | внутреннее соединение | implemented |
| auth-go -> postgres_auth | PostgreSQL | internal network | implemented |
| engine-go -> postgres_engine | PostgreSQL | internal network | implemented |
| BFF -> SMTP | SMTP/STARTTLS or SMTP | depends on provider config | implemented |

## Important Architectural Rules

1. В системе нет брокера сообщений.
2. Все ключевые бизнес-сценарии сейчас синхронные.
3. BFF является orchestration-layer, а не просто proxy.
4. Файл отчёта не хранится на диске сервера.
5. Email клиента должен присутствовать в `client_metadata`, иначе автоматическая отправка отчёта невозможна.
6. Bootstrap admin создаётся на старте `auth-go` из env, а не через hardcoded SQL secret.

## Current Technical Note

В текущем compose:

- `bff-go -> auth-go` и `bff-go -> engine-go` идут по `mTLS`;
- `bff-go -> analytics-python` пока настроен как внутренний `gRPC` без `mTLS`.

То есть целевая схема платформы предполагает защищённые service-to-service соединения везде, но фактическая текущая реализация Python analytics ещё требует отдельного доведения до полного `mTLS`.

## Recommended Presentation Version

Если нужно показать схему на защите или команде, используй именно этот короткий verbal summary:

> Frontend ходит только в BFF.  
> BFF оркестрирует `auth-go`, `engine-go` и `analytics-python`.  
> `auth-go` управляет входом, инвайтами и жизненным циклом аккаунтов.  
> `engine-go` хранит тесты, сессии и ответы.  
> `analytics-python` собирает отчёт.  
> Готовый отчёт отправляется клиенту на email и не сохраняется на сервере.
