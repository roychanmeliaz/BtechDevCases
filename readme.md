# Auth & Wallet Management System

A robust authentication and wallet management system built with Go, featuring JWT-based authentication, session management, and money transfer capabilities. Designed to handle poor network conditions with idempotency support.

## Features

- **User Authentication**
  - Registration with email validation
  - Login with JWT token generation
  - Password hashing with bcrypt
  
- **Session Management**
  - Redis-based session tracking
  - Automatic logout after 15 minutes of inactivity
  - Activity-based session renewal

- **Wallet System**
  - Automatic wallet creation on registration
  - Initial balance of 1000 units
  - View balance and transaction history
  - Transfer money to other users
  - Idempotent transfers for network reliability

- **Security**
  - JWT-based authentication
  - Password strength validation (minimum 8 characters)
  - Secure password storage with bcrypt
  - SQL injection prevention via parameterized queries

## Tech Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **ORM**: GORM
- **Authentication**: JWT (golang-jwt/jwt)

## Project Structure

```
.
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/    # HTTP request handlers
│   │   ├── middleware/  # Authentication middleware
│   │   └── router.go    # Route definitions
│   ├── config/          # Configuration management
│   ├── models/          # Database models
│   ├── repository/      # Database operations
│   └── service/         # Business logic
├── pkg/
│   └── jwt/             # JWT utilities
├── docker-compose.yml   # Docker services configuration
├── Dockerfile           # Application container
└── README.md
```

## Prerequisites

- Docker and Docker Compose
- OR: Go 1.21+, PostgreSQL 15, Redis 7

## Quick Start with Docker

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd BtechDevCases
   ```

2. **Start all services**
   ```bash
   docker-compose up --build
   ```

3. **Access the application**
   - API: http://localhost:8080
   - Health check: http://localhost:8080/health

## Manual Setup (Without Docker)

1. **Install dependencies**
   ```bash
   go mod download
   ```

2. **Start PostgreSQL and Redis**
   ```bash
   # PostgreSQL
   docker run -d --name postgres -p 5432:5432 \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=postgres \
     -e POSTGRES_DB=authwallet \
     postgres:15-alpine

   # Redis
   docker run -d --name redis -p 6379:6379 redis:7-alpine
   ```

3. **Configure environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run the application**
   ```bash
   go run cmd/server/main.go
   ```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HOST` | Server host address | 0.0.0.0 |
| `SERVER_PORT` | Server port | 8080 |
| `DB_HOST` | PostgreSQL host | localhost |
| `DB_PORT` | PostgreSQL port | 5432 |
| `DB_USER` | Database user | postgres |
| `DB_PASSWORD` | Database password | postgres |
| `DB_NAME` | Database name | authwallet |
| `DB_SSLMODE` | SSL mode | disable |
| `REDIS_HOST` | Redis host | localhost |
| `REDIS_PORT` | Redis port | 6379 |
| `REDIS_PASSWORD` | Redis password | "" |
| `REDIS_DB` | Redis database number | 0 |
| `JWT_SECRET` | JWT signing secret | (required in production) |
| `JWT_ACCESS_EXPIRATION_HOURS` | Token expiration | 24 |
| `SESSION_TIMEOUT_MINUTES` | Inactivity timeout | 15 |

## API Documentation

### Base URL
```
http://localhost:8080/api
```

### Endpoints

#### 1. Register
Create a new user account.

**Endpoint:** `POST /api/auth/register`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123",
  "confirmPassword": "securePassword123"
}
```

**Response:** `201 Created`
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
- `400` - Validation error (password mismatch, weak password)
- `409` - Email already exists

---

#### 2. Login
Authenticate and receive JWT token.

**Endpoint:** `POST /api/auth/login`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Response:** `200 OK`
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com"
  }
}
```

**Error Responses:**
- `400` - Invalid request format
- `401` - Invalid credentials

---

#### 3. Get Current User (Protected)
Get current user information.

**Endpoint:** `GET /api/me`

**Headers:**
```
Authorization: Bearer <jwt-token>
```

**Response:** `200 OK`
```json
{
  "message": "Hello user@example.com, welcome back",
  "user_id": 1
}
```

**Error Responses:**
- `401` - Unauthorized (invalid/expired token or session)

---

#### 4. Get Wallet (Protected)
View wallet balance and transaction history.

**Endpoint:** `GET /api/wallet`

**Headers:**
```
Authorization: Bearer <jwt-token>
```

**Response:** `200 OK`
```json
{
  "wallet": {
    "id": 1,
    "user_id": 1,
    "balance": 1000.0,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "transactions": [
    {
      "id": 1,
      "wallet_id": 1,
      "amount": 100.0,
      "type": "debit",
      "related_user_id": 2,
      "notes": "Payment for services",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `401` - Unauthorized
- `500` - Internal server error

---

#### 5. Transfer Money (Protected)
Transfer money to another user.

**Endpoint:** `POST /api/wallet/transfer`

**Headers:**
```
Authorization: Bearer <jwt-token>
Idempotency-Key: <unique-key> (optional, recommended for poor networks)
```

**Request Body:**
```json
{
  "recipient": "recipient@example.com",
  "amount": 100.0,
  "notes": "Payment for services"
}
```

**Response:** `200 OK`
```json
{
  "message": "transfer successful"
}
```

**Error Responses:**
- `400` - Insufficient balance, invalid amount, or self-transfer
- `401` - Unauthorized
- `404` - Recipient not found
- `500` - Internal server error

---

## Usage Examples

### Registration and Login Flow

```bash
# 1. Register a new user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "password123",
    "confirmPassword": "password123"
  }'

# 2. Login
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "password123"
  }' | jq -r '.token')

# 3. Access protected endpoint
curl http://localhost:8080/api/me \
  -H "Authorization: Bearer $TOKEN"
```

### Wallet Operations

```bash
# View wallet
curl http://localhost:8080/api/wallet \
  -H "Authorization: Bearer $TOKEN"

# Transfer money (with idempotency key for poor network resilience)
curl -X POST http://localhost:8080/api/wallet/transfer \
  -H "Authorization: Bearer $TOKEN" \
  -H "Idempotency-Key: $(uuidgen)" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": "bob@example.com",
    "amount": 50.0,
    "notes": "Payment for coffee"
  }'
```

## Network Resilience

This application is designed to handle poor network conditions (e.g., jungle or caves):

1. **Idempotency Keys**: Use the `Idempotency-Key` header for transfer requests to prevent duplicate transactions due to network retries.

2. **Atomic Transactions**: All database operations use transactions to ensure data consistency.

3. **Session Management**: Redis-based sessions with automatic timeout handling.

4. **Connection Pooling**: Efficient database connection management via GORM.

## Development

### Running Tests
```bash
go test ./...
```

### Building the Application
```bash
go build -o server ./cmd/server
```

### Database Migrations
The application uses GORM AutoMigrate for database schema management. Tables are automatically created/updated on startup.

## Stopping the Application

### Docker
```bash
docker-compose down
```

### Manual
Press `Ctrl+C` to stop the application.

## Security Considerations

1. **Change JWT Secret**: Always set a strong `JWT_SECRET` in production.
2. **Use HTTPS**: Enable TLS/SSL in production environments.
3. **Database Security**: Use strong passwords and restrict database access.
4. **Redis Security**: Enable authentication for Redis in production.
5. **Rate Limiting**: Consider adding rate limiting for API endpoints.

## License

MIT License

## Support

For issues or questions, please open an issue in the repository.
