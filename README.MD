# GoOne — All-in-One Golang API Starter Kit

**GoOne** is a production-ready Go API starter kit built on clean architecture principles. Everything you need to build scalable, maintainable REST APIs is already wired up — just start coding your business logic.

> **All in one:** HTTP server, auth, database, migrations, logging, AWS integrations, OpenAI, Swagger docs, DI — all included and ready to use.

---

## Tech Stack

| Layer | Technology |
| --- | --- |
| HTTP Framework | [Echo v4](https://echo.labstack.com/) |
| ORM | [GORM](https://gorm.io/) |
| Database | PostgreSQL / MySQL / MariaDB / SQLite |
| Auth | JWT (RS256 / HS256) |
| Migrations | [Goose](https://github.com/pressly/goose) (SQL-based) |
| Dependency Injection | [Google Wire](https://github.com/google/wire) |
| Logging | [Uber Zap](https://github.com/uber-go/zap) (structured JSON) |
| API Docs | [Swaggo](https://github.com/swaggo/swag) (Swagger / OpenAPI) |
| AWS | SES · S3 · SNS · SQS |
| AI | [OpenAI Go SDK](https://github.com/openai/openai-go) |

---

## Project Structure

```
├── cmd/
│   ├── api/            # HTTP server entry point
│   └── migration/      # Database migration CLI
├── config/             # App configuration (env vars)
├── internal/
│   ├── api/
│   │   ├── docs/       # Auto-generated Swagger docs
│   │   ├── handler/    # HTTP handlers (per module)
│   │   ├── router/     # Route registration
│   │   └── service/    # Business logic + DTOs (per module)
│   ├── di/             # Wire DI providers & generated code
│   ├── migrations/     # SQL migration files (Goose)
│   ├── model/          # GORM models
│   └── repository/     # DB repository implementations
└── pkg/
    ├── aws/            # SES, S3, SNS, SQS wrappers
    ├── database/       # Base repository with CRUD
    ├── logging/        # Zap logger with context support
    ├── openai/         # OpenAI chat, audio, streaming
    ├── server/         # Echo setup, middleware, error handling
    └── util/           # Config, crypter, request helpers, migrations
```

---

## Getting Started

### Prerequisites

*   [Go](https://go.dev/dl/) 1.24+
*   [Docker](https://docs.docker.com/install/) & [Docker Compose](https://docs.docker.com/compose/install/)
*   [swaggo/swag](https://github.com/swaggo/swag#getting-started) — for generating API docs
*   [air](https://github.com/air-verse/air) — for hot reload in development
*   [wire](https://github.com/google/wire) — for regenerating DI code

### First-time Setup

```
# 1. Start the database and run migrations
make provision

# 2. Run the development server with hot reload
make dev
```

The API is available at `http://localhost:8080`.  
Swagger UI is available at `http://localhost:8080/docs/index.html`.

### Default Credentials

```
POST /login HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
    "username": "superadmin",
    "password": "superadmin123!@#"
}
```

Use the returned `access_token` as a Bearer token for authenticated requests:

```
GET /v1/users HTTP/1.1
Host: localhost:8080
Authorization: Bearer <access_token>
```

---

## Development Commands

```
make dev              # Run with hot reload (air)
make specs            # Regenerate Swagger docs
make wire             # Regenerate Wire DI code
make test.cover       # Run tests with coverage report

# Migrations
make migrate          # Run pending migrations
make migrate.status   # Show migration status
make migrate.undo     # Rollback last migration
make migrate.create name=<name>   # Create new migration file

# Docker
make docker.build     # Build production Docker image
make docker.run       # Run container
make docker.logs      # Tail container logs
make docker.stop      # Stop and remove container
```

---

## Architecture Pattern

GoOne follows a strict **Service → Handler → Router** layered pattern:

1.  **Repository** — data access, wraps GORM with custom queries
2.  **Service** — business logic, DTOs, interface definitions
3.  **Handler** — HTTP binding, delegates to service
4.  **Router** — registers routes and attaches middleware

Dependency injection is handled at compile-time via **Google Wire**, ensuring zero runtime reflection overhead.

---

## Configuration

Environment variables are loaded with this priority (highest → lowest):

1.  OS environment variables _(always wins — ideal for Docker/K8s)_
2.  `.env.local` _(local overrides, gitignored)_
3.  `.env` _(committed defaults)_

Key variables: `STAGE`, `PORT`, `DB_TYPE`, `DB_DSN`, `JWT_SECRET`, `JWT_DURATION`, `JWT_ALGORITHM`, `ALLOW_ORIGINS`.

---

## Docker Deployment

The included `Dockerfile` uses a multi-stage build producing a ~20–30MB image with auto-migration on startup.

```
make docker.build
make docker.run
```

Health check endpoint: `GET /health`

---

## Installation

```
go get -u github.com/kevinxvu/goone
```

---

## License

MIT — see [LICENSE](https://github.com/kevinxvu/goone/blob/main/LICENSE) for details.