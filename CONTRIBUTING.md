# Contributing to promptpet

Guidelines and instructions for developing, testing, and contributing to `promptpet`.

## Prerequisites

- **Go:** Version `1.26` or later.
- **Git:** Standard git client.
- **Shells (for full integration testing):** `bash`, `zsh`, `fish`.
- **GoReleaser (optional, for release verification):** `v2+`.

## Development Workflow

### 1. Clone & Setup

```sh
git clone https://github.com/dpologdvinity/promptpet.git
cd promptpet
go mod download
go mod verify
```

### 2. Build

Compile both local development binaries (`promptpet` and the `promptpet-gui` window):

```sh
go build -o bin/ . ./cmd/promptpet-gui
```

Run CLI from source:

```sh
go run .
```

### 3. Testing

Run all unit and integration tests:

```sh
go test ./...
```

Run tests with data race detection:

```sh
go test -race ./...
```

Run tests with coverage profiling:

```sh
go test -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -func=coverage.out
```

### 4. Benchmarking

The prompt generation command must execute in under 10 ms. The in-process benchmark measures the code; the latency test builds the real binary and times it, which also catches slow package init (CI runs it on Linux; on a busy machine it can fail from noise):

```sh
go test -bench BenchmarkPrompt -benchmem ./internal/cli/
PROMPTPET_LATENCY_TEST=1 go test -count=1 -run Latency -v ./test/
```

### 5. Formatting & Code Quality

Format all Go source files:

```sh
gofmt -s -w .
```

Verify formatting compliance:

```sh
test -z "$(gofmt -l .)"
```

Run Go static analysis:

```sh
go vet ./...
```

---

## Commit Guidelines

This project strictly follows the **Conventional Commits** specification:

- `feat:` Introduces a new feature or capability.
- `fix:` Patches a bug or regression.
- `docs:` Documentation updates.
- `test:` Adds or refactors unit/integration tests.
- `refactor:` Code changes that neither fix bugs nor add features.
- `perf:` Performance improvements.
- `chore:` Tooling, workflow, build configuration, or repository maintenance.

### Commit Rules

1. Write concise, imperative commit headers (e.g., `feat(cli): add prompt styling flags`).
2. Keep commits atomic and logically separated.
3. Ensure `go test ./...` and `go vet ./...` pass before every commit.
4. Never commit personal agent config (`.claude/settings.local.json`, `CLAUDE.local.md`, `.superpowers/`). Shared agent config (`AGENTS.md`, `.claude/settings.json`, `docs/agents/`) is committed. See [AGENTS.md](AGENTS.md).

### Git Hooks

[lefthook](https://lefthook.dev) runs `gitleaks`, `gofmt` and `typos` on staged files before each commit. Enable it with:

```sh
lefthook install
```

---

## Pull Request Process

1. Create a feature branch off `master`.
2. Ensure all tests, race detectors, and benchmarks pass locally.
3. Verify formatting with `gofmt -s -l .`.
4. Submit a pull request targeting `master`.
5. Ensure all automated GitHub Actions CI status checks pass.
