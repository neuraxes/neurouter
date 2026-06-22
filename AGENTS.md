# AGENTS.md

Guidance for agents working in this repository. Keep changes consistent with the conventions below.

## Project Overview

Neurouter (`github.com/neuraxes/neurouter`) is an LLM router / proxy written in Go. It exposes multiple client-facing API formats and routes requests to upstream LLM providers.

- **Client APIs**: native gRPC (port 9000), native HTTP/REST (port 8000), plus OpenAI-, Anthropic- (Claude Code), and Ollama-compatible REST layers on the same HTTP server.
- **Upstream providers**: OpenAI (and OpenAI-compatible services), Anthropic, Google Gemini, and chained Neurouter instances.
- **Core features**: rate limiting (at upstream and model level), Probe-Rank-Reserve model election with load balancing, optional JWT auth (`JWT_KEY` env var), OpenTelemetry tracing/metrics/logging, Prometheus `/metrics`.

See `README.md` for user-facing docs.

## Tech Stack

- **Go 1.26.0**, Protocol Buffers 3
- **Kratos v2** (`github.com/go-kratos/kratos/v2`) app framework, HTTP + gRPC transports
- **Google Wire** for compile-time dependency injection
- Upstream SDKs: `openai-go/v3`, `anthropic-sdk-go`, `google.golang.org/genai`
- Observability: OpenTelemetry + `prometheus/client_golang`
- Tests: **GoConvey** (`github.com/smartystreets/goconvey`)

There is no buf, Taskfile, or Node.js tooling. Code generation uses raw `protoc` via the `Makefile`.

## Commands

Use the `Makefile` (run `make help` to list targets):

```bash
make init       # install protoc plugins, kratos CLI, and wire (run once)
make api        # generate api/*.proto -> *.pb.go, *_grpc.pb.go, *_http.pb.go, errors, openapi.yaml
make config     # generate internal/*.proto -> *.pb.go
make generate   # go generate ./... (regenerates Wire) + go mod tidy
make all        # api + config + generate
make build      # CGO_ENABLED=0 build -> bin/neurouter
```

Test and run:

```bash
go test ./... -v                 # run full test suite
./bin/neurouter -conf configs    # run locally after build
```

## Architecture

Standard Kratos layered layout. All application code lives under `internal/` (not externally importable).

```
api/neurouter/v1/   Public API: .proto sources + generated *.pb.go + hand-written helpers
cmd/neurouter/      main.go + Wire injectors (wire.go / wire_gen.go)
configs/            Runtime YAML configs
internal/
  service/          Thin gRPC handlers (RouterService) delegating to biz
  server/           HTTP/gRPC server setup + compat handlers (openai/, anthropic/, ollama/)
  biz/              Business logic (chat/, embedding/, model/, entity/, repository/)
  data/             Infrastructure: upstream/ providers, limiter/, telemetry/
  conf/             Internal config protos (conf.proto, upstream.proto)
  util/             Shared helpers
third_party/        Vendored proto deps (google/api, errors, validate, openapi)
```

Request flow:

```
Client -> internal/server (HTTP/gRPC or compat layer)
       -> internal/service.RouterService
       -> internal/biz (chat / embedding / model use cases)
       -> internal/data/upstream/* (provider SDK or HTTP)
```

Key concepts:

- **`internal/biz/repository`** holds interfaces only (e.g. `ChatRepo`, `EmbeddingRepo`, `Limiter`, factories).
- **Model election** (`internal/biz/model/election.go`): Probe (check delay across limiters) -> Rank (shuffle available, sort waitable by delay) -> Reserve (reserve limiters, wait, or fall back).
- **Rate limiters** (`internal/data/limiter/`): token bucket, daily quota, concurrency semaphore, composed at upstream and model level.

## Where to Add Code

- **New upstream provider** → `internal/data/upstream/<name>/`, register in the data `ProviderSet`.
- **New compatibility API** → `internal/server/<name>/` (transport/protocol adapter only, no business logic).
- **New domain logic** → `internal/biz/<area>/`; expose interfaces via `internal/biz/repository`.
- **New API surface** → edit the `.proto` in `api/neurouter/v1/`, then run `make all`.
- After adding any provider/use case, wire it into the relevant `ProviderSet` and run `make generate`.

## Code Conventions

- All code, comments, identifiers, commit messages, documentation, and CI output in this repository **must be written in clear, correct English**.
- Documentation describes the **current** state. Do not keep "previously we did ..." / "compared to the old design" notes; history lives in git.
- **The code is the design of record.** There is no external design document; the code and its doc comments are the single source of truth for both _what_ the system does and _why_. Never point a reader at a file outside the code (or a section number in one) — inline what they need.
- **Make names do the explaining.** Every type, struct/interface, field and function name must reflect its actual purpose, so that reading the code is normally enough to understand it. Fix an unclear name before adding a comment to compensate for it.
- **Prefer no comment when the code already speaks.** A comment that merely restates what the code plainly does is noise — omit it. Comment only what code cannot carry: non-obvious _why_, trade-offs, invariants, edge cases, and constraints.

- Every Go source file starts with the Apache 2.0 header. Copy it from an existing file when creating new ones.
- Import grouping: standard library, then third-party, then `github.com/neuraxes/neurouter/...`.
- Packages are short and lowercase (`chat`, `model`, `openai`, `local`).
- Interfaces use names like `UseCase`, `ChatRepo`, `Elector`; implementations are unexported structs (`chatUseCase`, `upstream`).
- Each package exposes a `ProviderSet` for Wire.
- Conversion between native API types and provider formats lives in `convert.go` / `*_convert.go`.
- Streaming uses Go iterators: `iter.Seq2[T, error]` with `for resp, err := range ...`.
- Startup failures in `main.go` use `panic(err)`. Business errors use Kratos errors generated from `error_reason.proto` (e.g. `v1.ErrorNoUpstream(...)`, `v1.ErrorTokenQuotaExhausted(...)`).
- Upstream factory failures are logged and skipped (`log.Errorf(...); continue`), not fatal.

## Proto / Codegen Rules

- **Never hand-edit generated files** (`*.pb.go`, `*_grpc.pb.go`, `*_http.pb.go`, `*_errors.pb.go`, `openapi.yaml`, `wire_gen.go`). Edit the `.proto`/`wire.go` source and regenerate with `make`.
- Hand-written extension methods on generated types go in plain `.go` files in the same package (see `api/neurouter/v1/meta.go`, `text.go`, `json.go`).
- Proto package naming: public API is `neurouter.v1`; internal config is `neurouter.config.v1`.
- Enums use `UPPER_SNAKE_CASE` with a type prefix (`CHAT_IN_PROGRESS`, `MODALITY_TEXT`, `CAPABILITY_CHAT`, `REASONING_EFFORT_HIGH`).
- Field-number conventions: `metadata` maps at field 15; `oneof` branches use high numbers (templates 50, grammar 60+, provider config 100+).
- Errors are defined in `error_reason.proto` using the Kratos `errors` extension; generated constructors are `v1.Error<ReasonCamelCase>()`.

## Testing

- Use **GoConvey** exclusively. Import with the dot form and nest `Convey` blocks:

```go
import . "github.com/smartystreets/goconvey/convey"

func TestSomething(t *testing.T) {
    Convey("Given ...", t, func() {
        Convey("when ...", func() {
            So(result, ShouldEqual, expected)
        })
    })
}
```

- Compare protobufs with `google.golang.org/protobuf/proto.Equal`, not `ShouldEqual`.
- Shared fixtures live in `mock_test.go` (e.g. `mockChatReq`, `mockChatStreamResp`); reuse them.
- Upstream integration tests mock the HTTP client (e.g. `mockHTTPClient` with a `DoFunc`).
- Conversion tests (`*_convert_test.go`) should cover each role/modality branch.
