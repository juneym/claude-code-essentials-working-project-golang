# SPECS.md — Hello World Tasks API

## Overview

A production-idiomatic Go HTTP API with two endpoints, full OpenAPI documentation served via Swagger UI, and an `AGENTS.md` describing the tech stack and dev conventions.

---

## Tech Stack

| Component       | Choice                                              | Version    |
|-----------------|-----------------------------------------------------|------------|
| Language        | Go                                                  | 1.22+      |
| HTTP Framework  | `github.com/gin-gonic/gin`                          | v1.10.x    |
| OpenAPI codegen | `github.com/swaggo/swag` (CLI + lib)                | v1.16.x    |
| Swagger UI      | `github.com/swaggo/gin-swagger` + `swaggo/files`    | v1.6.x / v1.0.x |
| System memory   | `github.com/shirou/gopsutil/v3/mem`                 | v3.24.x    |

---

## Endpoints

### `GET /echo` and `POST /echo`

Accepts a message and echoes it back unchanged.

- **GET**: message passed as query param `?message=<value>`
- **POST**: message passed as JSON body `{"message": "<value>"}`
- **Validation**: `message` is required; returns HTTP 400 with `ErrorResponse` if absent
- **Response**: `{"echo": "<value>"}`

### `GET /stats`

Returns current server and runtime statistics.

**Response shape:**
```json
{
  "server_time_utc": "2026-04-10T20:57:31Z",
  "sys_mem": {
    "total_bytes": 8589934592,
    "used_bytes": 6840385536,
    "used_percent": 79.63
  },
  "go_mem": {
    "heap_alloc_bytes": 10546464,
    "heap_in_use_bytes": 12476416,
    "total_alloc_bytes": 12086792
  }
}
```

- `sys_mem` sourced from `gopsutil/v3/mem.VirtualMemory()` — cross-platform (macOS + Linux)
- `go_mem` sourced from `runtime.ReadMemStats()` — triggers a brief stop-the-world GC pause (~10–50µs); acceptable for an infrequent monitoring endpoint

### `GET /swagger/index.html`

Serves interactive Swagger UI from the generated `docs/` package.

---

## Project Structure

```
hello-world-tasks/
├── AGENTS.md                        # tech stack, install guide, conventions, dev workflow
├── SPECS.md                         # this file
├── go.mod / go.sum                  # pinned dependencies
├── main.go                          # Gin engine, routes, Swagger UI, swag top-level annotations
├── docs/                            # generated OpenAPI spec — never hand-edit
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/
│   └── handler/
│       ├── handler.go               # shared ErrorResponse type + RegisterRoutes()
│       ├── echo.go                  # EchoHandler with swag annotations
│       ├── stats.go                 # StatsHandler with swag annotations
│       ├── echo_test.go             # unit tests for EchoHandler (package handler)
│       ├── stats_test.go            # unit tests for StatsHandler (package handler)
│       └── api_test.go              # black-box API integration tests (package handler_test)
└── scripts/
    └── swag-init.sh                 # convenience wrapper: swag init -g main.go --output docs
```

---

## Architecture Decisions

### Why Gin over stdlib `net/http`
`swaggo/swag` has first-class Gin support. `ShouldBind` handles both GET (query/form) and POST (JSON body) in a single handler via content-type negotiation. Gin's `Logger` and `Recovery` middleware are production-appropriate defaults.

### Why `swaggo/swag` over `go-swagger`
Go code is the source of truth, not a separate YAML spec. `swag init` is a single CLI command that regenerates `docs/` from annotations. CI drift-detection runs `swag init` + `git diff --exit-code docs/`.

### Why `gopsutil` over `/proc/meminfo`
`/proc/meminfo` is Linux-only. `gopsutil` works correctly on macOS (Darwin) and Linux without any platform-specific parsing logic.

### `docs/` in version control
The generated `docs/` package must be committed because `main.go` has a blank import `_ "github.com/juneym/hello-world-tasks/docs"`. The package must exist at build time. Teams preferring clean VCS can instead run `swag init` as a `go generate` step in CI before building.

### Echo handler: single function for GET + POST
`ShouldBind` inspects the `Content-Type` header and resolves from query/form tags (GET) or JSON body (POST) automatically. Both verbs are registered to the same `EchoHandler` function; only one set of swag `@Router` annotations per operation is needed.

---

## Testing

### Strategy

Tests are co-located with the handler package under `internal/handler/` and split into two layers:

| Layer | File | Go package | Approach |
|---|---|---|---|
| Unit | `echo_test.go` | `package handler` | White-box; calls handlers directly via `httptest.NewRecorder` |
| Unit | `stats_test.go` | `package handler` | White-box; calls handlers directly via `httptest.NewRecorder` |
| API / Integration | `api_test.go` | `package handler_test` | Black-box; calls `RegisterRoutes()` and exercises the full router |

No external test server is started. All tests use `net/http/httptest` — the full HTTP dispatch path runs in-process, making the suite fast and self-contained.

---

### Unit Tests — Echo Handler (`echo_test.go`)

| Test | Description |
|---|---|
| `TestEchoHandler_GET_ReturnsMessage` | GET with `?message=hello+world` returns `{"echo":"hello world"}` and HTTP 200 |
| `TestEchoHandler_GET_SpecialCharacters` | URL-encoded emoji and punctuation are decoded and echoed correctly |
| `TestEchoHandler_GET_MissingMessage_Returns400` | GET with no `message` param returns HTTP 400 with non-empty `error` field |
| `TestEchoHandler_POST_JSONBody` | POST with `Content-Type: application/json` and `{"message":"..."}` returns HTTP 200 and echoed value |
| `TestEchoHandler_POST_EmptyBody_Returns400` | POST with `{}` (missing required field) returns HTTP 400 |
| `TestEchoHandler_POST_NoContentType_Returns400` | POST without `Content-Type` header falls back to form binding; no `message` → HTTP 400 |
| `TestEchoHandler_ResponseContentType` | Response `Content-Type` header contains `application/json` |

---

### Unit Tests — Stats Handler (`stats_test.go`)

| Test | Description |
|---|---|
| `TestStatsHandler_Returns200` | GET `/stats` returns HTTP 200 |
| `TestStatsHandler_ResponseHasAllFields` | All seven fields (`server_time_utc`, `sys_mem.*`, `go_mem.*`) are present and non-zero |
| `TestStatsHandler_ServerTimeIsRFC3339UTC` | `server_time_utc` parses as RFC3339, location is UTC, and falls within the test execution window |
| `TestStatsHandler_SysMemUsedLessThanTotal` | `used_bytes ≤ total_bytes` and `used_percent ≤ 100` |
| `TestStatsHandler_ResponseContentType` | Response `Content-Type` header contains `application/json` |

---

### API Integration Tests (`api_test.go`)

| Test | Description |
|---|---|
| `TestAPI_Echo_GET_Returns200WithEcho` | Full router: GET `/echo?message=api+test` → `{"echo":"api test"}`, HTTP 200 |
| `TestAPI_Echo_POST_Returns200WithEcho` | Full router: POST `/echo` with JSON body → echoed value, HTTP 200 |
| `TestAPI_Echo_GET_MissingMessage_Returns400` | Full router: GET `/echo` → HTTP 400, `error` key present in body |
| `TestAPI_Echo_POST_MissingMessage_Returns400` | Full router: POST `/echo` with `{}` → HTTP 400, `error` key present |
| `TestAPI_Echo_MessageIsPreservedVerbatim` | Message with spaces, tabs, and newlines is echoed character-for-character |
| `TestAPI_Stats_Returns200` | Full router: GET `/stats` → HTTP 200, `Content-Type: application/json` |
| `TestAPI_Stats_ResponseStructure` | Top-level keys `server_time_utc`, `sys_mem`, `go_mem` present; nested keys verified |
| `TestAPI_Stats_ServerTimeIsRecentUTC` | `server_time_utc` is valid RFC3339, falls within test execution window |
| `TestAPI_Stats_MemoryValuesArePositive` | All numeric memory fields are `> 0` |
| `TestAPI_UnknownRoute_Returns404` | Request to unknown path returns HTTP 404 |

---

### Running Tests

```sh
# All tests
go test ./...

# Verbose output
go test ./internal/handler/... -v

# With coverage
go test ./internal/handler/... -coverprofile=coverage.out
go tool cover -func=coverage.out

# Single test by name
go test ./internal/handler/... -run TestAPI_Stats
```

---

## Key Implementation Notes

- `docs/` stub must exist before `go mod tidy` so the local import resolves; `swag init` then overwrites it with the real spec.
- `runtime.ReadMemStats` is a stop-the-world call. For high-frequency polling of `/stats`, introduce a 1-second refresh cache backed by a ticker goroutine.
- `PORT` env var (default `8080`) and `GIN_MODE` env var (default `debug`, set to `release` in production).

---

## Running the Application

```sh
# Install Go (macOS)
brew install go

# Install swag CLI (one-time)
go install github.com/swaggo/swag/cmd/swag@latest

# Resolve dependencies
go mod tidy

# Regenerate OpenAPI docs (after annotation changes)
swag init -g main.go --output docs
# or:
./scripts/swag-init.sh

# Start server
go run ./main.go

# Build binary
go build -o hello-world-tasks .
```

---

## Verification

```sh
# Echo GET
curl -s "http://localhost:8080/echo?message=hello+world"
# → {"echo":"hello world"}

# Echo POST
curl -s -X POST http://localhost:8080/echo \
  -H "Content-Type: application/json" \
  -d '{"message":"hello from POST"}'
# → {"echo":"hello from POST"}

# Echo — missing message (400)
curl -s "http://localhost:8080/echo"
# → {"error":"Key: 'EchoRequest.Message' Error:Field validation for 'Message' failed on the 'required' tag"}

# Stats
curl -s "http://localhost:8080/stats"
# → {"server_time_utc":"...","sys_mem":{...},"go_mem":{...}}

# Swagger UI (open in browser)
open http://localhost:8080/swagger/index.html
```
