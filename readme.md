# Take-home Assignment: Auth with JWT (TypeScript)

Build a small application in **TypeScript/Go/C#** that supports user **registration** and **login** using **JWT**.
You can choose any stack or structure you want.
As long as the core flow works end-to-end, it's accepted.
Please note User will be using the app in place with very bad connections, like jungle or caves.

---

## Requirements

### 1. Register

- Fields: `email`, `password`, `confirmPassword`

### 2. Login

- Input: `email`, `password`
- Return: **JWT**
- Token should contain at least:
  - `email`
  - `user id` or similar identifier

### 3. Authenticated View / Endpoint

After successful login, calling the protected route / loading the protected screen should show:

```
Hello [email], welcome back
```

user should be logged out after 15 minutes of inacitvity

---

### 4. Manager wallet

User should be able to see and transfer his money to other user.
fields are: recipient, amount, and notes

## What to deliver

- Fork this repository and then send the link
- A runnable project (any structure).
- README explaining:
  - How to build and run it (prepare docker compose)
  - Required environment variables

---

## Acceptance criteria

- Registration works with validation.
- Login returns a usable JWT.
- A protected route or screen shows the welcome message using JWT auth.
- User able to transfer funds

---

## Optional bonus

- Docker
- Backend built using Go (or their frameworks)
- Frontend built using React/Vue (or their frameworks)
- Tests (unit or integration)

This keeps the scope tight: just registration, login, and a protected "Hello [email]" flow.

---

---

# My Implementation

I am using **Go** with the Gin framework. Go performance will be a good fit for the requirements, including to facilitae the poor network conditions mentioned above.

## What I Built

### Core Features
- User registration with email/password validation
- Login that returns a JWT token (with user_id and email in claims)
- Protected `/api/me` endpoint that shows "Hello [email], welcome back"
- 15-minute inactivity timeout using Redis sessions
- Wallet system where users can view balance and transfer money
- Full transaction history

### Extra Touches
I added a few things beyond the requirements:
- **Idempotent transfers**: Since users might be in areas with bad connections, I implemented idempotency keys to prevent duplicate transactions on retry
- **Docker setup**: Everything runs with a single `docker-compose up` command
- **Unit tests**: Added tests for the core business logic (78.6% coverage on services)
- **Double-entry bookkeeping**: Proper transaction recording with debit/credit entries
- **Initial wallet balance**: New users start with 1000 units to test transfers right away

## Tech Stack

I chose this stack:
- **Go** with Gin framework
- **PostgreSQL** for data persistence
- **Redis** for session management
- **GORM** as the ORM
- **JWT** for authentication

## How to Run

### Docker
Just run this:
```bash
docker-compose up --build
```

Open `http://localhost:8080/health` to verify it's running.

### Manual Setup
Running without Docker:

1. Make sure you have Go 1.23+, PostgreSQL, and Redis installed
2. Start PostgreSQL and Redis (or use the Docker commands below)
3. Run `go mod download` to install dependencies
4. Run `go run cmd/server/main.go`

## Project Structure

```
cmd/server/        → Main application
internal/
  ├── api/         → HTTP handlers, middleware, router
  ├── config/      → Environment config
  ├── models/      → Database models
  ├── repository/  → Data access layer
  └── service/     → Business logic
pkg/jwt/           → Reusable JWT utilities
```

## Configuration

There's a `.env.example` file to copy. No need to setup if using Docker Compose.

Key settings:
- `JWT_SECRET` - Change in production
- `SESSION_TIMEOUT_MINUTES` - Set to 15 as required
- Database and Redis connection settings

Check `.env.example` for the full list.

## API Endpoints

All endpoints are under `/api`. Here's what I implemented:

### 1. Register a User
`POST /api/auth/register`

Creates a new user and automatically creates a wallet with 1000 units.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "confirmPassword": "password123"
}
```

**Success Response (201):**
```json
{
  "message": "registration successful",
  "user": {
    "id": 1,
    "email": "user@example.com"
  }
}
```

**Error Responses:**
- `400` - Password mismatch or weak password
- `409` - Email already exists

### 2. Login
`POST /api/auth/login`

Returns a JWT token that you'll use for authenticated requests.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Success Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6InVzZXJAZXhhbXBsZS5jb20iLCJleHAiOjE3MDk5MjgwMDAsIm5iZiI6MTcwOTg0MTYwMCwiaWF0IjoxNzA5ODQxNjAwfQ.xyz123",
  "user": {
    "id": 1,
    "email": "user@example.com"
  }
}
```

**Error Responses:**
- `400` - Invalid request format
- `401` - Invalid email or password

### 3. Get Current User (Protected)
`GET /api/me`

Returns the "Hello [email], welcome back" message. Requires JWT token in Authorization header.

**Request Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**Success Response (200):**
```json
{
  "message": "Hello user@example.com, welcome back",
  "user_id": 1
}
```

**Error Responses:**
- `401` - Missing or invalid token
- `401` - Session expired (15 minutes of inactivity)

### 4. View Wallet
`GET /api/wallet`

Shows your balance and transaction history.

**Request Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**Success Response (200):**
```json
{
  "wallet": {
    "id": 1,
    "user_id": 1,
    "balance": 849.5,
    "created_at": "2026-02-11T14:22:27.873616Z",
    "updated_at": "2026-02-11T14:23:52.262112Z"
  },
  "transactions": [
    {
      "id": 1,
      "wallet_id": 1,
      "amount": 150.5,
      "type": "debit",
      "related_user_id": 2,
      "notes": "Payment for coffee",
      "created_at": "2026-02-11T14:23:52.26349Z",
      "updated_at": "2026-02-11T14:23:52.26349Z"
    }
  ]
}
```

**Error Responses:**
- `401` - Unauthorized
- `500` - Internal server error

### 5. Transfer Money
`POST /api/wallet/transfer`

Use the `Idempotency-Key` header to prevent duplicate transfers on network retries.

**Request Headers:**
```
Authorization: Bearer <your-jwt-token>
Idempotency-Key: unique-request-id-123  (optional but recommended)
```

**Request Body:**
```json
{
  "recipient": "bob@example.com",
  "amount": 150.50,
  "notes": "Coffee payment"
}
```

**Success Response (200):**
```json
{
  "message": "transfer successful"
}
```

**Error Responses:**
- `400` - Insufficient balance
- `400` - Invalid amount (must be greater than 0)
- `400` - Cannot transfer to yourself
- `401` - Unauthorized
- `404` - Recipient not found

## Quick Test

Here's the quick flow:

```bash
# 1. Register
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","confirmPassword":"password123"}'

# 2. Login and save token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' | jq -r '.token')

# 3. Access protected route
curl http://localhost:8080/api/me \
  -H "Authorization: Bearer $TOKEN"
# Should return: "Hello test@example.com, welcome back"

# 4. Check wallet
curl http://localhost:8080/api/wallet \
  -H "Authorization: Bearer $TOKEN"
```

## Testing

I wrote unit tests for the core business logic:

```bash
go test ./...              # Run all tests
go test ./... -cover       # With coverage
```

Coverage:
- Service layer: **78.6%**
- JWT utilities: **88.9%**

Tests cover:
- Registration/login flows
- Wallet operations and transfers
- Idempotency checks
- Error handling
- JWT validation

## Design Decisions

### Why Idempotency?
The brief mentioned users in caves/jungles with bad connections. Idempotency keys ensure that if a user's request times out and they retry, we won't process the transfer twice.

### Why Redis for Sessions?
I needed to track the 15-minute inactivity timeout. Redis is perfect for this - it has built-in TTL (time-to-live) and we can reset it on each request.

### Why GORM?
It handles migrations automatically and provides a clean API. Tables are created on startup, so you don't need to run migrations manually.

## To Add

- Rate limiting (especially for login attempts)
- Email verification
- More comprehensive integration tests
- Prometheus metrics
- Better logging (structured logs with correlation IDs)
- Admin endpoints for user management

## Notes

- The initial wallet balance (1000) is just for testing convenience
- JWT secret has a default value for dev
- Database schema auto-migrates on startup
- Sessions expire after 15 minutes of inactivity as required
