# Repository Guidelines

## Project Structure & Module Organization

Olivine is a Redis-compatible key-value service written in Go. The executable entrypoint lives in `cmd/olivine`, with application wiring in `app.go`, `kessoku.go`, and generated DI output in `kessoku_band.go`.

Core packages are under `internal/`: `internal/server` handles network serving and request dispatch, `internal/service` contains command and AOF logic, `internal/service/cmd` implements Redis commands, `internal/repo` stores data, and `internal/data` parses configuration. RESP protocol parsing and value encoding live in `pkg/resp`. Tests are colocated with source as `*_test.go`. Runtime configuration examples are in `redis.conf`.

## Build, Test, and Development Commands

- `make build`: builds `cmd/olivine` into `bin/olivine`.
- `make test`: runs `go test ./...` across all packages.
- `make gen`: runs `go generate ./...`; use this after changing generated DI or mock inputs.
- `make clean`: removes the local `bin` directory.
- `go run ./cmd/olivine`: runs the server locally during development.

Prefer `make` targets when they exist so local and CI-style workflows stay aligned.

## Coding Style & Naming Conventions

Use standard Go formatting: run `gofmt` on edited Go files. Keep package names short and lowercase, and follow Go naming conventions for exported identifiers. Tests should use clear behavior names and table-driven cases where helpful. Generated files such as `kessoku_band.go` and `storage_mock.go` should be updated through `go generate`, not hand-edited.

## Testing Guidelines

The project uses Go's built-in testing package, with `go-cmp` where structural comparisons are useful. Add tests next to the package being changed, using `TestXxx` names and descriptive subtests. For command behavior, follow the existing patterns in `internal/service/cmd/*_test.go`; for RESP parsing and encoding, use `pkg/resp/*_test.go`. Run `make test` before opening a PR.

## Commit & Pull Request Guidelines

Recent history uses concise Conventional Commit-style messages such as `fix: handle keepttl for AOF` and `docs: add license`. Keep commits focused and describe the user-visible behavior or internal subsystem changed.

Pull requests should include a short summary, test results such as `make test`, and any linked issue or tutorial chapter when relevant. For protocol or persistence changes, mention compatibility impact and include representative command examples.

## Agent-Specific Instructions

Do not overwrite generated files manually. If an interface, provider graph, or mock source changes, run `make gen` and then `make test`. Keep responses and documentation concise, but preserve precise file paths and commands.
