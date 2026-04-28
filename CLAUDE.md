# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test

```bash
# Run all tests
go test -v ./...

# Run tests for a specific package
go test -v ./kratos/caching/...
go test -v ./kratos/orm/...

# Generate protobuf code (requires protoc)
make api
```

Module: `github.com/omalloc/contrib` (Go 1.24).

## Architecture

This is a Go monorepo of reusable packages for the [Kratos](https://go-kratos.dev/) microservice framework. Every package is self-contained — no package imports another package within this repo (except `protobuf` and `net/broadcast` which are imported by their sibling packages).

### `kratos/` — Kratos framework integrations

- **`kratos/orm`** — GORM wrapper. `orm.New()` creates a `*gorm.DB` with functional options for driver, tracing (OpenTelemetry plugin), and logger. Subpackage `orm/crud` provides a generic `CRUD[T]` interface for Create/Update/Delete/SelectList/SelectOne. `transaction.go` provides context-based transaction propagation via `context.Context` value keys — use `NewTransactionManager` to get a `Transaction` that stores the `*gorm.DB` tx in context.
- **`kratos/caching`** — Generic `LoadableCache[K, V]` backed by gcache with periodic auto-refresh. Supports retry, blocking initial load, and OpenTelemetry tracing. Use `New[K, V](opts...)` with `WithRefreshAfterWrite` to define the data loader function.
- **`kratos/registry`** — Etcd-based `Registrar` and `Discovery` for Kratos, configured via `protobuf.Registry`. Returns `nil` registrar when `OnlyDiscovery` is set.
- **`kratos/gin`** — Adapts Kratos HTTP middleware chain (`[]middleware.Middleware`) into a `gin.HandlerFunc`. Also provides `Error()` to render Kratos errors as JSON responses.
- **`kratos/resty`** — go-resty HTTP client with OpenTelemetry tracing via custom `http.Transport` round-tripper.
- **`kratos/tracing`** — Jaeger tracer initialization. Call `InitTracer(opts...)` once at startup; configures the global OTel tracer provider.
- **`kratos/zap`** — Uber zap logger wrapper. `New()` returns a production config tuned for Kratos (suppressed time/caller keys to avoid duplication). `NewLogger()` wraps it as a `kratoszap.Logger`.
- **`kratos/health`** — Health check HTTP service. Register `Checker` implementations; exposes `/health` endpoint returning component status (UP/DOWN).
- **`kratos/selector/filter`** — Kratos node selector filter that excludes nodes marked with metadata `hang=true`.

### `x/` — Standalone utilities

- **`x/singleflight`** — Generics-based duplicate function call suppression (`Group[K, V]`). Fork of `golang.org/x/sync/singleflight` with Go 1.18+ generics.
- **`x/duration`** — `Duration` type (wraps `time.Duration`) with JSON marshal/unmarshal as human-readable strings like `"5s"`.
- **`x/runtime`** — `AutoGOMAXPROCS()` for CPU/memory tuning, `PrintStackTrace()` for crash diagnostics, and `SetCurrentUser()`/`ParseUser()` for dropping privileges.

### Other top-level packages

- **`net/broadcast`** — UDP broadcast-based service discovery protocol. Shared message types/encoding in the root `broadcast` package, with `client/` (query sender) and `server/` (responder). Server listens on a discovery port (default `5353`), responds with the service's advertised address when queried.
- **`machineid`** — Machine fingerprinting using CPU info, product UUID, and default gateway MAC. `Load().SN()` returns a stable machine identifier.
- **`protobuf`** — Shared protobuf messages: `Pagination` (with GORM `Paginate()` scope, offset/limit helpers) and `Registry` (etcd config).
- **`magic`** — Example/demo package showing the `Magic` interface pattern.

## Code conventions

- Functional options pattern is used pervasively: each constructor takes `...Option` where `Option` is `func(*Config)`.
- Generics are used where applicable (`cache[K, V]`, `CRUD[T]`, `Group[K, V]`).
- Packages are flat — avoid deeply nested directory structures.
