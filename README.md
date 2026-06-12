# lumbera-hackathon

A production-ready Go REST API boilerplate built on a clean 3-layer architecture. Clone this template to skip the initial setup and jump straight into writing business logic.

## Tech Stack

| Package                                                       | Purpose                     |
| ------------------------------------------------------------- | --------------------------- |
| [Gin](https://github.com/gin-gonic/gin)                       | HTTP web framework          |
| [GORM](https://gorm.io)                                       | ORM for database operations |
| [gorm/driver/mysql](https://github.com/go-gorm/mysql)         | MariaDB/MySQL driver        |
| [golang-jwt/jwt](https://github.com/golang-jwt/jwt)           | JWT authentication          |
| [google/uuid](https://github.com/google/uuid)                 | UUID generation             |
| [joho/godotenv](https://github.com/joho/godotenv)             | `.env` file loading         |
| [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) | Bcrypt password hashing     |

---

## Folder Structure

```
golang-project-template/
├── cmd/
│   └── app/
│       └── main.go           # Entry point: wires all dependencies and starts the server
├── entity/
│   └── your_entity.go        # GORM models (mapped to database tables)
├── internal/                 # Core application code (3-layer architecture)
│   ├── handler/
│   │   └── rest/
│   │       └── rest.go       # HTTP layer: receives requests, sends responses
│   ├── repository/
│   │   └── repository.go     # Data access layer: database operations via GORM
│   └── service/
│       └── service.go        # Business logic layer
├── model/
│   └── your_model.go         # DTOs for request/response serialization
├── pkg/                      # Shared utilities
│   ├── bcrypt/
│   │   └── bcrypt.go         # Password hashing (bcrypt, cost=10)
│   ├── config/
│   │   ├── config.go         # Loads .env file via godotenv
│   │   └── database.go       # Builds the DSN connection string
│   ├── constant/
│   │   └── role.go           # Application-wide constants (e.g. role UUIDs)
│   ├── database/
│   │   ├── mariadb.go        # Opens the GORM database connection
│   │   └── migrate.go        # Runs GORM AutoMigrate on startup
│   ├── errors/
│   │   └── errors.go         # Custom AppError type with HTTP status codes
│   ├── jwt/
│   │   └── jwt.go            # JWT creation and validation
│   ├── middleware/
│   │   └── middleware.go     # Gin middleware (auth guards, etc.)
│   └── response/
│       └── response.go       # Standardized JSON response envelope
├── .env.example              # Environment variable template
├── go.mod
└── go.sum
```

---

## Architecture

The core of this template lives in the `internal/` directory, which enforces a strict 3-layer separation of concerns.

```
HTTP Request
     │
     ▼
┌─────────────────────┐
│   Handler / REST    │  ← Receives requests, validates input, returns responses
│  internal/handler/  │
└────────┬────────────┘
         │ calls
         ▼
┌─────────────────────┐
│      Service        │  ← Business logic, orchestrates data flow
│  internal/service/  │
└────────┬────────────┘
         │ calls
         ▼
┌─────────────────────┐
│     Repository      │  ← Database operations (GORM queries)
│ internal/repository/│
└────────┬────────────┘
         │
         ▼
      Database
    (MariaDB / MySQL)
```

### Layer Responsibilities

**`internal/repository/`**
Owns all direct database interaction. Each method corresponds to a specific query or mutation. Receives a `*gorm.DB` instance and is the only layer allowed to call GORM methods.

**`internal/service/`**
Contains business logic. Depends on the repository for data access and on `pkg/bcrypt` and `pkg/jwt` for cross-cutting concerns. Never imports GORM directly.

**`internal/handler/rest/`**
The outermost layer. Binds HTTP routes via Gin, parses request bodies, calls the service, and writes back JSON responses using the shared `pkg/response` formatter.

---

## Dependency Injection

All dependencies are wired manually in `cmd/app/main.go` using constructor injection — no DI framework required.

```
main()
  ├── config.LoadEnvironment()         # Load .env
  ├── mariadb.ConnectDatabase()        # Open *gorm.DB
  ├── mariadb.Migrate()                # Auto-migrate tables
  │
  ├── repository.NewRepository(db)     # Data layer
  ├── bcrypt.Init()                    # Password util
  ├── jwt.Init()                       # Auth util
  ├── service.NewService(repo, bcrypt, jwt)   # Business logic
  ├── middleware.Init(service, jwt)    # Middleware chain
  └── rest.NewRest(service, middleware)
        ├── rest.MountEndpoint()       # Register routes
        └── rest.Run()                 # Start server
```

---

## Environment Variables

Copy `.env.example` to `.env` and fill in the values before running.

| Variable         | Description                               | Example              |
| ---------------- | ----------------------------------------- | -------------------- |
| `DB_HOST`        | Database host                             | `localhost`          |
| `DB_PORT`        | Database port                             | `3306`               |
| `DB_NAME`        | Database name                             | `myapp`              |
| `DB_USER`        | Database user                             | `root`               |
| `DB_PASSWORD`    | Database password                         | `secret`             |
| `ADDRESS`        | Server bind address                       | `localhost`          |
| `PORT`           | Server port                               | `8080`               |
| `TIME_OUT_LIMIT` | Request timeout (seconds)                 | `10`                 |
| `JWT_SECRET_KEY` | Secret key for signing JWTs (min 256-bit) | `a-very-long-secret` |
| `JWT_EXP_TIME`   | JWT expiration in hours                   | `1`                  |

---

## Getting Started with This Template

1. **Clone or use as template**

   ```bash
   git clone https://github.com/your-username/golang-project-template.git my-new-project
   cd my-new-project
   ```

2. **Update the module name**

   ```bash
   # Replace the module name in go.mod and all imports
   find . -type f -name "*.go" | xargs sed -i 's|golang-project-template|my-new-project|g'
   # Then update go.mod manually
   ```

3. **Set up environment**

   ```bash
   cp .env.example .env
   # Edit .env with your database credentials and config
   ```

4. **Install dependencies**

   ```bash
   go mod tidy
   ```

5. **Start building**
   - Define your database models in `entity/`
   - Add them to `pkg/database/migrate.go`
   - Add request/response structs in `model/`
   - Implement repository methods in `internal/repository/`
   - Implement business logic in `internal/service/`
   - Register routes and handlers in `internal/handler/rest/`

6. **Run**

   ```bash
   go run cmd/app/main.go
   ```

   ```bash
   air
   ```
