# URL Shortener

A production-oriented URL shortener built in Go.

This project is being developed as a systems-design exercise. V1 focuses on establishing a clean, testable, maintainable core before introducing distributed-system complexity such as Redis, replication, asynchronous processing, and horizontal scaling.

## V1.0.0

V1 provides the complete core URL-shortening workflow:

- Validate a user-provided URL
- Generate a cryptographically secure short code
- Persist the URL mapping in PostgreSQL
- Retrieve URLs using their short code
- Redirect users to the original URL
- Gracefully shut down the HTTP server
- Test the generator, validator, repository, handlers, and integration flow

The project is intentionally kept simple at this stage.

The goal of V1 is to establish a correct and extensible foundation rather than prematurely optimize for large-scale traffic.

---

## Features

### URL Creation

Create a short URL by submitting an HTTP/HTTPS URL.

`POST /api/v1/urls`

Example request:

```json
{
  "url": "https://github.com/"
}
```

Example response:

```json
{
  "short_code": "aB3xYz1"
}
```

The generated short code is currently 7 characters long.

### URL Redirection

A generated short code can be accessed directly:

`GET /{shortCode}`

For example:

`GET /aB3xYz1`

The server looks up the short code in PostgreSQL and responds with:

`302 Found`

redirecting the client to the original URL.

### URL Validation

Only valid HTTP and HTTPS URLs are accepted.

The validator currently checks:

- URL is present
- URL does not exceed 2048 characters
- URL can be parsed
- Scheme is either `http` or `https`
- Host is present

Invalid requests are rejected with:

`400 Bad Request`

### Collision Protection

Generated short codes are stored in a database column with a `UNIQUE` constraint.

This means the database remains the final authority on short-code uniqueness.

The generator produces random short codes, while PostgreSQL guarantees that duplicate values cannot be persisted.

The current V1 design does not automatically retry after a collision. Collision retry can be introduced later when the persistence layer evolves.

### Graceful Shutdown

The HTTP server listens for termination signals and attempts to gracefully shut down.

Active requests are given a short window to complete before the server exits.

This avoids abruptly terminating the server during normal shutdown.

---

## Architecture

The project follows a layered architecture.

```text
                    ┌────────────────────┐
                    │      HTTP Client   │
                    └─────────┬──────────┘
                              │
                              v
                    ┌────────────────────┐
                    │      Handler       │
                    │                    │
                    │ - Decode request   │
                    │ - Validate URL     │
                    │ - Generate code    │
                    └─────────┬──────────┘
                              │
                              v
                    ┌────────────────────┐
                    │    Repository      │
                    │                    │
                    │ - Create URL       │
                    │ - Get by shortCode │
                    └─────────┬──────────┘
                              │
                              v
                    ┌────────────────────┐
                    │    PostgreSQL      │
                    └────────────────────┘
```

Supporting components:

```text
config
  |
  +---- Application configuration

database
  |
  +---- PostgreSQL connection pool

generator
  |
  +---- Short-code generation

validator
  |
  +---- URL validation

repository
  |
  +---- Persistence abstraction

handler
  |
  +---- HTTP/API layer
```

The separation is deliberate.

The HTTP layer does not directly know how PostgreSQL works.

The repository layer does not know about HTTP requests.

The generator does not know about persistence.

This makes each component independently testable and easier to replace later.

---

## Project Structure

```text
url-shortener/
│
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── config/
│   │   └── config.go
│   │
│   ├── database/
│   │   └── database.go
│   │
│   ├── generator/
│   │   ├── generator.go
│   │   └── generator_test.go
│   │
│   ├── handler/
│   │   ├── health.go
│   │   ├── integration_test.go
│   │   ├── url.go
│   │   └── url_test.go
│   │
│   ├── repository/
│   │   ├── repository.go
│   │   └── postgres/
│   │       ├── postgres.go
│   │       └── postgres_test.go
│   │
│   └── validator/
│       ├── url.go
│       └── url_test.go
│
├── migrations/
│   └── 001_create_urls.sql
│
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

---

## Technology Stack

### Backend

- Go
- `net/http`

The project uses Go's standard HTTP library instead of introducing a web framework.

This keeps the HTTP layer explicit and minimizes framework-specific abstractions.

### Database

- PostgreSQL
- `pgx/v5`

PostgreSQL is currently the source of truth for URL mappings.

The application uses `pgxpool` for connection pooling.

### Testing

The project uses Go's standard testing package:

`testing`

The test suite covers:

- Short-code generation
- Generator validation
- Generated character validation
- Short-code uniqueness
- URL validation
- PostgreSQL repository operations
- Duplicate short-code handling
- Handler behavior
- Redirect behavior
- Integration flow

---

# Setup

## Requirements

Before running the application, install:

- Go 1.26.3 or a compatible newer Go version
- PostgreSQL

Check your Go installation:

```bash
go version
```

Check PostgreSQL:

```bash
psql --version
```

## 1. Clone the repository

```bash
git clone https://github.com/Ashwanijha1405/url-shortener.git
cd url-shortener
```

## 2. Create the PostgreSQL database

Create a database named:

`url_shortener`

Using `psql`:

```bash
psql -U postgres
```

Then:

```sql
CREATE DATABASE url_shortener;
```

Exit:

```sql
\q
```

## 3. Run the database migration

Connect to the database:

```bash
psql -U postgres -d url_shortener
```

Run the migration:

```sql
CREATE TABLE IF NOT EXISTS urls (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

The same schema is available in:

`migrations/001_create_urls.sql`

You can also execute the migration directly:

```bash
psql -U postgres -d url_shortener -f migrations/001_create_urls.sql
```

## 4. Configure the database connection

The application reads the PostgreSQL connection string from:

`DATABASE_URL`

Create a local `.env` file based on `.env.example`.

Example:

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/url_shortener
PORT=8080
```

### Important

`.env` is intentionally ignored by Git.

Never commit credentials or other secrets to the repository.

The application currently reads environment variables directly. Therefore, `.env` does not automatically become part of the process environment unless a dotenv loader is introduced.

### PowerShell

```powershell
$env:DATABASE_URL="postgres://postgres:postgres@localhost:5432/url_shortener"
$env:PORT="8080"
```

### Linux/macOS

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/url_shortener"
export PORT="8080"
```

---

# Running the Application

From the project root:

```bash
go run ./cmd/server
```

The server starts on:

`http://localhost:8080`

The port can be changed using the `PORT` environment variable.

For example:

```powershell
$env:PORT="9000"
go run ./cmd/server
```

---

# API

## Health Check

`GET /health`

Example:

```bash
curl http://localhost:8080/health
```

This endpoint is intended for basic service health verification and future deployment/load-balancer health checks.

## Create a Short URL

### Endpoint

`POST /api/v1/urls`

### Request

```json
{
  "url": "https://github.com/"
}
```

Example using curl in PowerShell:

```powershell
curl.exe -X POST http://localhost:8080/api/v1/urls `
  -H "Content-Type: application/json" `
  -d "{\"url\":\"https://github.com/\"}"
```

Example response:

```json
{
  "short_code": "aB3xYz1"
}
```

HTTP status:

`201 Created`

## Redirect

After receiving a short code:

`GET /aB3xYz1`

The server performs:

```text
GET /aB3xYz1
      |
      v
Database lookup
      |
      v
Original URL
      |
      v
302 Found
      |
      v
Original destination
```

Example:

```powershell
curl.exe -I http://localhost:8080/aB3xYz1
```

Expected response:

```text
HTTP/1.1 302 Found
Location: https://github.com/
```

---

# Error Handling

The API currently returns standard HTTP error responses.

### Invalid request body

`400 Bad Request`

### Missing URL

`400 Bad Request`

### Invalid URL

`400 Bad Request`

### Short code not found

`404 Not Found`

### Database or internal failure

`500 Internal Server Error`

The API intentionally does not expose internal database errors directly to clients.

Internal errors are wrapped and handled inside the application while clients receive a safe HTTP-level error.

---

# Database Design

V1 uses a single table:

`urls`

Schema:

```text
┌──────────────┬────────────────────────────┐
│ Column       │ Type                       │
├──────────────┼────────────────────────────┤
│ id           │ BIGSERIAL                  │
│ short_code   │ VARCHAR(10) UNIQUE        │
│ original_url │ TEXT                       │
│ created_at   │ TIMESTAMPTZ                │
└──────────────┴────────────────────────────┘
```

The important constraint is:

```sql
UNIQUE(short_code)
```

The database is responsible for enforcing uniqueness.

This is preferable to relying exclusively on an application-level check such as:

```text
SELECT first
INSERT later
```

because concurrent requests could otherwise race between the check and insert.

---

# Short-Code Generation

The current generator uses a 62-character alphabet:

`0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`

The default length is:

`7 characters`

The generator uses:

`crypto/rand`

rather than a predictable pseudo-random generator.

The current design therefore prioritizes unpredictable short-code generation over deterministic sequential IDs.

## Why 7 Characters?

With 62 possible characters and 7 positions:

`62^7 ≈ 3.5 trillion`

possible combinations exist.

This is more than sufficient for the scope of V1.

However, the number of possible combinations does not mean collisions are impossible.

Therefore:

```text
Random generation
        +
Database UNIQUE constraint
        =
Correctness
```

As the system grows, collision handling can be improved with retries and more explicit error classification.

---

# Repository Abstraction

The handler depends on an interface rather than directly depending on PostgreSQL.

Conceptually:

```go
type URLRepository interface {
    Create(
        ctx context.Context,
        shortCode string,
        originalURL string,
    ) error

    GetByShortCode(
        ctx context.Context,
        shortCode string,
    ) (string, error)
}
```

The PostgreSQL implementation satisfies this interface.

This provides an important architectural boundary:

```text
Handler
   |
   v
URLRepository interface
   |
   +---- PostgreSQL implementation
   |
   +---- Future mock/fake implementation
   |
   +---- Future alternative storage
```

The handler therefore does not need to change if the persistence implementation changes.

---

# Connection Pooling

The database layer uses `pgxpool`.

The application maintains a pool of PostgreSQL connections instead of opening a new connection for every HTTP request.

This is important because URL shorteners are expected to be read-heavy systems.

A connection pool provides:

- Connection reuse
- Lower connection establishment overhead
- Bounded database connections
- Better behavior under concurrent requests

The current V1 configuration keeps the pool intentionally conservative.

It can be tuned later based on actual workload and deployment characteristics.

---

# Testing

Run the complete test suite:

```bash
go test ./... -v
```

The repository integration tests require a running PostgreSQL instance and a correctly configured:

`DATABASE_URL`

## Test Categories

### Generator Tests

Verify:

- Default length
- Custom length
- Invalid lengths
- Valid characters
- Uniqueness across multiple generated codes

### Validator Tests

Verify:

- Valid HTTPS URLs
- Valid HTTP URLs
- Empty URLs
- Unsupported schemes
- Missing hosts
- Invalid URLs

### Repository Tests

Verify:

- URL creation
- URL retrieval
- Missing short codes
- Duplicate short codes

These tests interact with PostgreSQL because repository behavior depends on database constraints and queries.

### Handler Tests

Verify:

- URL creation
- Invalid requests
- Invalid JSON
- Repository errors
- Redirect behavior
- Missing short codes

### Integration Tests

The integration test verifies the complete workflow:

```text
HTTP Request
     |
     v
Handler
     |
     v
Validator
     |
     v
Generator
     |
     v
Repository
     |
     v
PostgreSQL
```

and then:

```text
Short Code
     |
     v
Redirect Request
     |
     v
Repository
     |
     v
Original URL
     |
     v
HTTP 302
```

This is important because unit tests alone cannot guarantee that all layers work correctly together.

---

# Design Decisions

## Standard Library HTTP Server

V1 uses:

`net/http`

instead of a third-party web framework.

### Reason

The project is intended to explore backend and system-design fundamentals.

Using the standard library keeps:

- Routing explicit
- Dependencies minimal
- HTTP behavior visible
- The learning surface smaller

A framework can be introduced later if the project develops requirements that justify it.

## Repository Pattern

Database access is isolated behind a repository interface.

### Reason

This prevents business logic from becoming tightly coupled to PostgreSQL.

It also makes testing easier and leaves room for future storage changes.

## PostgreSQL as the Source of Truth

V1 intentionally has no cache.

Every redirect lookup goes through PostgreSQL.

### Reason

Correctness and simplicity are more important at this stage.

Adding Redis before measuring the actual bottleneck would introduce additional infrastructure and consistency concerns without solving a demonstrated problem.

## Random Short Codes

V1 uses cryptographically secure random generation.

### Reason

It avoids predictable sequential identifiers and keeps the implementation simple.

A distributed ID-generation strategy is intentionally postponed.

## Database-Level Uniqueness

The application generates candidate short codes, but PostgreSQL enforces uniqueness.

### Reason

The database must protect invariants even when multiple application instances operate concurrently.

## Explicit API Versioning

The creation endpoint is:

`/api/v1/urls`

### Reason

Introducing the version boundary early makes future API evolution easier.

Future incompatible API changes can be introduced under:

`/api/v2/...`

without immediately breaking V1 clients.

---

# Current Limitations

V1 intentionally does not solve every production-scale problem.

Current limitations include:

- Single PostgreSQL instance
- No Redis cache
- No database replication
- No asynchronous processing
- No analytics
- No rate limiting
- No authentication
- No custom aliases
- No URL expiration
- No click tracking
- No abuse detection
- No distributed ID generation
- No horizontal deployment configuration
- No automated migration runner
- No structured application logging
- No metrics/tracing system
- No retry policy for generated-code collisions

These are deliberate scope boundaries rather than accidental omissions.

---

# Roadmap

## V1 — Core URL Shortener

Completed:

- [x] Project structure
- [x] PostgreSQL integration
- [x] Database schema
- [x] Repository abstraction
- [x] Short-code generator
- [x] URL validation
- [x] URL creation API
- [x] Redirect API
- [x] Health endpoint
- [x] Unit tests
- [x] Repository tests
- [x] Handler tests
- [x] Integration tests
- [x] Environment-based configuration
- [x] Graceful shutdown

## V2 — Performance and Reliability

Potential additions:

- Redis caching
- Cache-aside strategy
- Collision retry handling
- Better database indexing
- Structured logging
- Metrics
- Request IDs
- More robust error classification
- Rate limiting

Expected redirect flow:

```text
                 ┌─────────────┐
                 │    Client   │
                 └──────┬──────┘
                        │
                        v
                 ┌─────────────┐
                 │   Handler   │
                 └──────┬──────┘
                        │
                        v
                 ┌─────────────┐
                 │    Redis    │
                 └──────┬──────┘
                        │
              cache miss│
                        v
                 ┌─────────────┐
                 │ PostgreSQL  │
                 └─────────────┘
```

The cache should only be introduced once the system has a clear read-heavy workload and the consistency model is understood.

## V3 — Scale the Database Layer

Potential additions:

- Read replicas
- Primary/replica routing
- Connection pool tuning
- Database partitioning where justified
- Improved migration management

The main design question will become:

`How do we scale reads without compromising correctness?`

## V4 — Distributed ID Generation

The random short-code generator can eventually be replaced or supplemented with a distributed ID-generation strategy.

Possible approaches include:

- Snowflake-style IDs
- Database-generated numeric IDs followed by Base62 encoding
- Dedicated ID-generation service

The choice should be driven by:

- Uniqueness
- Ordering
- Coordination
- Throughput
- Length
- Operational complexity

There is no need to introduce distributed ID generation while V1 remains a single-service deployment.

## V5 — Analytics

Potential additions:

- Click count
- Timestamp of access
- Referrer
- User-agent information
- Approximate geographic information
- Analytics API

Analytics should not unnecessarily block the redirect path.

A likely future architecture is:

```text
Client
  |
  v
Redirect Service
  |
  +-------> URL lookup
  |
  +-------> Event
                |
                v
          Message Queue
                |
                v
          Analytics Worker
                |
                v
          Analytics Store
```

This allows the user-facing redirect operation to remain fast while analytics are processed asynchronously.

---

# Future System Architecture

The eventual architecture may evolve toward:

```text
                         ┌───────────────┐
                         │    Clients    │
                         └───────┬───────┘
                                 │
                                 v
                         ┌───────────────┐
                         │ Load Balancer │
                         └───────┬───────┘
                                 │
                    ┌────────────┼────────────┐
                    │            │            │
                    v            v            v
                ┌───────┐    ┌───────┐    ┌───────┐
                │ App 1 │    │ App 2 │    │ App N │
                └───┬───┘    └───┬───┘    └───┬───┘
                    │            │            │
                    └────────────┼────────────┘
                                 │
                         ┌───────┴───────┐
                         │     Redis     │
                         └───────┬───────┘
                                 │
                           Cache miss
                                 │
                                 v
                         ┌───────────────┐
                         │   PostgreSQL  │
                         │    Primary    │
                         └───────┬───────┘
                                 │
                         Replication
                                 │
                    ┌────────────┴────────────┐
                    v                         v
             ┌───────────────┐        ┌───────────────┐
             │   Read Replica│        │   Read Replica│
             └───────────────┘        └───────────────┘
```

Analytics can later be introduced asynchronously without coupling it to the redirect request.

The important principle is that each layer should be added because the workload or reliability requirements justify it.

---

# Engineering Principles

This project follows a few principles throughout its development.

### 1. Correctness before optimization

A simple correct system is preferable to a complex system whose behavior is difficult to reason about.

### 2. Introduce infrastructure only when justified

Redis, message queues, replicas, distributed IDs, and other infrastructure should solve actual problems rather than exist merely because they are common in system-design diagrams.

### 3. Keep boundaries explicit

HTTP handling, validation, generation, persistence, and configuration should remain independently understandable.

### 4. Let the database enforce database invariants

Application code can generate candidate values, but constraints such as uniqueness belong at the database level.

### 5. Optimize the architecture incrementally

The system is intentionally designed so that future improvements can be introduced without rewriting the entire application.

---

# Development Workflow

For normal development:

```bash
git pull origin main
```

Create a feature branch:

```bash
git checkout -b feature/<name>
```

Run formatting:

```bash
gofmt -w .
```

Run tests:

```bash
go test ./... -v
```

Review changes:

```bash
git diff
git status
```

Commit:

```bash
git add .
git commit -m "your commit message"
```

Push:

```bash
git push origin feature/<name>
```

Open a pull request against:

`main`

---

# Release Philosophy

The project uses incremental releases.

A release should represent a coherent set of capabilities rather than an arbitrary collection of features.

V1 establishes the foundation.

Future versions should preserve existing behavior wherever possible and introduce new complexity only when the architectural problem is clearly defined.

The long-term goal is not simply to build a URL shortener.

The goal is to evolve the same codebase through progressively more demanding system-design problems while keeping each stage understandable, testable, and maintainable.

---

# License

This project is currently provided for educational and experimentation purposes.

License information can be added when the project is formally licensed for redistribution.

---

# Author

**Ashwani Jha**

GitHub:  
https://github.com/Ashwanijha1405

Repository:  
https://github.com/Ashwanijha1405/url-shortener
