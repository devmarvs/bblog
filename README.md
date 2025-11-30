# bblog API

bblog is a REST API for logging baby (or pet) activities. It lets a primary user create sub-users, capture activity logs, and review history through authenticated endpoints.
> **Status:** Work in progress – expect active development and frequent changes.

## Tech Stack
- Go 1.23 with Gin for the web framework
- PostgreSQL for persistence (schema seeded automatically on startup)
- JWT for authentication and authorization
- bcrypt password hashing
- Docker/Docker Compose for containerized deployment

## Getting Started
1. **Prerequisites**
   - Go 1.23 or newer (for local builds)
   - PostgreSQL instance
   - `git`, `make`, and optionally Docker/Docker Compose

2. **Environment variables** (see `.env` for an example):
   - `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PWORD`, `DB_NAME`
   - `JWT_SECRET`
   - `GIN_MODE` (`debug` by default if unset)
   - `PORT` (defaults to `8080`)

3. **Run locally**
   ```bash
   go mod download
   go run ./...
   ```
   The API starts on `http://localhost:8080` (or the port defined in `PORT`).

4. **Run with Docker Compose**
   ```bash
   docker compose up --build
   ```
   Ensure `.env` points to a reachable PostgreSQL database. The migrate logic in `db.InitDb()` creates the `bblog` schema and seeds lookup tables on boot.

## API Overview
Base path: `http://localhost:8080/bblog`

| Method | Endpoint                                       | Auth | Purpose                          |
| ------ | ---------------------------------------------- | ---- | -------------------------------- |
| POST   | `/user/create`                                 | No   | Register a primary user          |
| POST   | `/login`                                       | No   | Authenticate and get JWT         |
| GET    | `/user/all`                                    | Yes  | List active users you can view   |
| GET    | `/user/{id}`                                   | Yes  | Fetch a user by id               |
| POST   | `/user/{id}/subuser`                           | Yes  | Create a sub-user (baby/pet)     |
| GET    | `/user/{id}/subuser`                           | Yes  | List sub-users owned by user     |
| POST   | `/subuser/log`                                 | Yes  | Record an activity log           |
| GET    | `/user/{id}/subuser/{subUserId}/log`           | Yes  | View logs for a specific sub-user|
| POST   | `/user/verify`                                 | No   | Verify email with a code         |
| GET    | `/user/verify`                                 | No   | Alternate verify using query     |
| POST   | `/user/resend-verification`                    | No   | Send another verification code   |
| POST   | `/user/verify-email`                           | No   | Preferred resend endpoint (POST) |
| GET    | `/user/verify-email`                           | No   | Resend endpoint (GET with query) |
| POST   | `/user/forgot-password`                        | No   | Email password reset link        |
| POST   | `/user/reset-password`                         | No   | Set new password via token       |
| GET    | `/user/types`                                  | Yes  | List supported user types        |
| GET    | `/log/types`                                   | Yes  | List supported log categories    |
| POST   | `/logout`                                      | Yes  | Revoke the current JWT           |

### Authentication
Add an `Authorization: Bearer <jwt>` header to any endpoint marked with **Auth = Yes**. Tokens expire two hours after issuance. Calling `/bblog/logout` revokes the presented token immediately so it cannot be reused even if time remains.

---

## Endpoint Details

### POST `/bblog/user/create`
Create a primary user. Passwords are stored as bcrypt hashes.

```http
POST /bblog/user/create
Content-Type: application/json

{
  "user_type_id": 1,
  "username": "parent01",
  "password": "StrongP@ssw0rd",
  "email": "parent@example.com",
  "mobile": "+15005550000",
  "country_code": "US"
}
```

**201 Created**
```json
{
  "data": null,
  "message": "User created. Check your email for the verification code."
}
```

The response body omits the new `user_id`. Capture it by logging in (next endpoint) or querying `/bblog/user/all` after the account is verified. New accounts remain inactive until the emailed verification code is confirmed. Codes expire after **5 minutes**; request a new code if it lapses.

---

### POST `/bblog/login`
Authenticate with email and password to receive a JWT.

```http
POST /bblog/login
Content-Type: application/json

{
  "email": "parent@example.com",
  "password": "StrongP@ssw0rd"
}
```

**200 OK**
```json
{
  "message": "Login Successful",
  "token": "<jwt-token>"
}
```

Use the returned token for subsequent protected requests. Attempts to log in before verifying the email return `403` with `"Please verify your email before logging in"`.

---

### POST `/bblog/user/verify`
Verify a user's email using the 6-digit code sent after registration. POST JSON (preferred) or supply `email` and `code` as query parameters (GET/POST both work).

```http
POST /bblog/user/verify
Content-Type: application/json

{
  "email": "parent@example.com",
  "code": "123456"
}
```

Query alternative:
```
GET /bblog/user/verify?email=parent@example.com&code=123456
```

**200 OK**
```json
{
  "message": "Email verified successfully. You can now log in."
}
```

- `400 Bad Request`: email/code missing, invalid, or expired (codes last 5 minutes).
- `409 Conflict`: code already used or account already verified.

After a successful verification the server marks the user as active and future login attempts succeed.

---

### POST `/bblog/user/resend-verification`
Legacy endpoint to trigger another verification code email. Use this if the first code expired or never arrived. Prefer `/bblog/user/verify-email` (below).

```http
POST /bblog/user/resend-verification
Content-Type: application/json

{
  "email": "parent@example.com"
}
```

**200 OK**
```json
{
  "message": "Verification code sent"
}
```

- `400 Bad Request`: email missing or the account is already verified.
- `500 Internal Server Error`: SMTP not configured or email could not be sent.

For privacy, the endpoint replies with `200` even if the email is unknown, but the message changes only when the account is already verified.

---

### POST `/bblog/user/verify-email`
Preferred resend endpoint for verification codes. Accepts the same payload as `/user/resend-verification`.

```http
POST /bblog/user/verify-email
Content-Type: application/json

{
  "email": "parent@example.com"
}
```

**200 OK**
```json
{
  "message": "Verification code sent"
}
```

- `400 Bad Request`: email missing or the account is already verified.
- `500 Internal Server Error`: SMTP not configured or email could not be sent.

You can also call `GET /bblog/user/verify-email?email=parent@example.com` without a body if that is easier for your client; it returns the same responses and does not require authentication.

---

### POST `/bblog/user/forgot-password`
Send a password reset link to the supplied email. The response is the same whether the account exists or not.

```http
POST /bblog/user/forgot-password
Content-Type: application/json

{
  "email": "parent@example.com"
}
```

**200 OK**
```json
{
  "message": "If the account exists, a reset email has been sent"
}
```

- `400 Bad Request`: email missing or invalid JSON payload.
- `500 Internal Server Error`: SMTP not configured or email could not be sent.

The link expires after one hour; request another email if it lapses.

---

### POST `/bblog/user/reset-password`
Complete the reset by posting the token from the email together with the new password.

```http
POST /bblog/user/reset-password
Content-Type: application/json

{
  "token": "<reset-token>",
  "password": "StrongerP@ssw0rd"
}
```

**200 OK**
```json
{
  "message": "Password updated successfully"
}
```

- `400 Bad Request`: token missing/invalid/expired or password too short.
- `409 Conflict`: token already used.
- `500 Internal Server Error`: unexpected error while updating the password.

Tokens are single-use. After a successful reset the user can log in immediately with the new password.

---

### POST `/bblog/logout` _(requires auth)_
Revoke the current session token. After a successful logout the same JWT can no longer access protected endpoints.

```http
POST /bblog/logout
Authorization: Bearer <jwt>
```

**200 OK**
```json
{
  "message": "Logout successful"
}
```

Call `/bblog/login` again to obtain a fresh token when needed.

---

### GET `/bblog/user/all` _(requires auth)_
Fetch all active, non-deleted users that the authenticated caller is permitted to view.

**200 OK**
```json
{
  "data": [
    {
      "user_id": 12,
      "username": "parent01",
      "created_ts": "2024-06-09T18:04:11Z",
      "user_type_id": 1,
      "email": "parent@example.com",
      "mobile": "+15005550000",
      "country_code": "US",
      "is_online": false,
      "is_active": true,
      "is_deleted": false,
      "is_premium": false
    }
  ]
}
```
If no users exist, `data` is `null`. Requests without a valid token receive `401`/`403`.

---

### GET `/bblog/user/{id}` _(requires auth)_
Retrieve a single user. You can only fetch your own user record.

```
Authorization: Bearer <jwt>
GET /bblog/user/12
```

**200 OK**
```json
{
  "data": {
    "user_id": 12,
    "username": "parent01",
    "created_ts": "2024-06-09T18:04:11Z",
    "user_type_id": 1,
    "email": "parent@example.com",
    "mobile": "+15005550000",
    "country_code": "US",
    "is_online": false,
    "is_active": true,
    "is_deleted": false,
    "is_premium": false
  }
}
```

Unauthorized or mismatched IDs return `401`/`403`.

---

### POST `/bblog/user/{id}/subuser` _(requires auth)_
Register a sub-user (child, baby, or pet) linked to the authenticated user.

```http
POST /bblog/user/12/subuser
Authorization: Bearer <jwt>
Content-Type: application/json

{
  "user_type_id": 2,
  "name": "Baby Finn"
}
```

**201 Created**
```json
{
  "data": {
    "sub_user_id": 44,
    "user_id": 12,
    "user_type_id": 2,
    "name": "Baby Finn",
    "is_active": true,
    "is_deleted": false,
    "created_ts": "2024-06-09T19:52:01Z",
    "updated_ts": ""
  }
}
```

---

### GET `/bblog/user/{id}/subuser` _(requires auth)_
List active sub-users owned by the authenticated parent user.

**200 OK**
```json
{
  "data": [
    {
      "sub_user_id": 44,
      "user_id": 12,
      "user_type_id": 2,
      "name": "Baby Finn",
      "is_active": true,
      "is_deleted": false,
      "created_ts": "2024-06-09T19:52:01Z",
      "updated_ts": ""
    }
  ]
}
```

---

### POST `/bblog/subuser/log` _(requires auth)_
Record a log entry for one of your sub-users. Valid `log_type_id` values are seeded in the `bblog.log_types` table (e.g., milk, diaper, temperature).

```http
POST /bblog/subuser/log
Authorization: Bearer <jwt>
Content-Type: application/json

{
  "sub_user_id": 44,
  "log_type_id": 1,
  "log_time": "2024-06-09T20:15:00Z",
  "log_description": "120ml of formula"
}
```

**201 Created**
```json
{
  "data": {
    "user_log": 133,
    "user_id": 12,
    "sub_user_id": 44,
    "log_type_id": 1,
    "log_time": "2024-06-09T20:15:00Z",
    "log_description": "120ml of formula",
    "created_ts": "2024-06-09T20:15:12.345Z"
  },
  "message": "User Log created Successfully"
}
```

---

### GET `/bblog/user/{id}/subuser/{subUserId}/log` _(requires auth)_
Retrieve log history for a specific sub-user, ordered newest first.

```http
GET /bblog/user/12/subuser/44/log
Authorization: Bearer <jwt>
```

**200 OK**
```json
{
  "data": [
    {
      "user_log": 133,
      "user_id": 12,
      "sub_user_id": 44,
      "name": "Baby Finn",
      "log_type_id": 1,
      "log_name": "milk",
      "log_time": "2024-06-09T20:15:00Z",
      "log_description": "120ml of formula",
      "created_ts": "2024-06-09T20:15:12.345Z",
      "updated_ts": ""
    }
  ]
}
```

---

### GET `/bblog/log/types` _(requires auth)_
Return the list of all activity categories you can log. Use this to populate client-side dropdowns or validation.

```http
GET /bblog/log/types
Authorization: Bearer <jwt>
```

**200 OK**
```json
{
  "data": [
    { "log_type_id": 1, "log_name": "milk" },
    { "log_type_id": 2, "log_name": "medicine" },
    { "log_type_id": 3, "log_name": "vaccine" }
    // ...
  ]
}
```

---

### GET `/bblog/user/types` _(requires auth)_
Return all available user roles (e.g., primary user, baby, pet).

```http
GET /bblog/user/types
Authorization: Bearer <jwt>
```

**200 OK**
```json
{
  "data": [
    { "user_type_id": 1, "description": "user" },
    { "user_type_id": 2, "description": "baby" },
    { "user_type_id": 3, "description": "pet" }
  ]
}
```

Use these IDs when creating users or sub-users.

---

## Error Handling
- `400 Bad Request`: malformed payloads or missing fields.
- `401 Unauthorized`: missing/invalid token.
- `403 Forbidden`: user tries to access resources they don’t own.
- `500 Internal Server Error`: database or server problems.

All error responses follow the pattern:
```json
{
  "data": null,
  "message": "Descriptive message",
  "error": "Optional error details"
}
```

## Contributing & Support
Feel free to fork the repository, open issues, or submit PRs. For questions, reach out via the project’s issue tracker.
