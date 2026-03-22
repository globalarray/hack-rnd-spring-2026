# Auth Service Integration

## Purpose

`auth-go` is the authentication and account lifecycle service for the platform.

It is responsible for:

- shared login entry point for administrators and psychologists
- issuing access and refresh tokens
- invitation-based onboarding for psychologists
- bootstrap admin creation on first launch
- account blocking and unblocking
- switching expired accounts to `inactive`

## Core Flow

### Administrator

1. Logs in through the shared endpoint.
2. Creates an invitation for a psychologist with:
   - `full_name`
   - `phone`
   - `email`
   - `access_until`
   - `expires_at`
3. Receives an invitation token and a frontend URL:
   - `https://example.com/invitations/{uuid}`
4. If the same psychologist needs a new link before registration, a new invitation replaces the previous unused invitation for the same email.

### Psychologist

1. Opens the invitation link.
2. Enters a password.
3. BFF calls `auth-go.Register`.
4. `auth-go` creates the psychologist account and issues tokens.

## Account Lifecycle

`users.status` supports:

- `active`
- `inactive`
- `blocked`

Rules:

- `blocked` is set manually by an administrator
- `inactive` is applied automatically when `access_until` is in the past
- expired users also lose their refresh token
- protected RPC methods validate not only JWT signature, but also the current user status in the database

## Bootstrap Administrator

On first startup `auth-go` checks whether the configured bootstrap admin already exists.

If not, it creates one automatically from env:

- `AUTH_BOOTSTRAP_ADMIN_EMAIL`
- `AUTH_BOOTSTRAP_ADMIN_PASSWORD`
- `AUTH_BOOTSTRAP_ADMIN_FULL_NAME`
- `AUTH_BOOTSTRAP_ADMIN_PHONE`
- `AUTH_BOOTSTRAP_ADMIN_ACCESS_UNTIL`
- `AUTH_BOOTSTRAP_ADMIN_ROLE`

This is implemented as startup seeding after schema initialization, not as a hardcoded SQL insert inside the migration.

Reason:

- admin credentials come from environment variables
- secrets should not be baked into SQL migrations

## Infrastructure

The stack now includes:

- `postgres_auth`
- `auth-go`

`auth-go` runs gRPC over mTLS and reuses the shared certificate bundle mounted to `/etc/certs`.

`bff-go` dials `auth-go` through internal gRPC and exposes frontend-ready REST endpoints.

## BFF Endpoints

Public:

- `POST /public/v1/auth/login`
- `POST /public/v1/auth/refresh`
- `POST /public/v1/auth/register`
- `GET /public/v1/profiles/{userId}`

Protected:

- `GET /api/v1/auth/profile`
- `PATCH /api/v1/auth/profile`
- `PATCH /api/v1/auth/users/{userId}/profile`
- `POST /api/v1/auth/invitations`
- `POST /api/v1/auth/users/block`
- `POST /api/v1/auth/users/unblock`

## Notes

- the shared login endpoint returns `role`, so the frontend can route admin and psychologist into different cabinets
- the invitation URL is assembled in BFF using `PUBLIC_BASE_URL`
- `expiresAt` in REST may be sent as `YYYY-MM-DD`; BFF expands it to the end of that day
- psychologist profile modification is allowed only through the administrator token via `PATCH /api/v1/auth/users/{userId}/profile`
- survey flow for candidates remains under `/public/v1/sessions/...` and does not require auth
