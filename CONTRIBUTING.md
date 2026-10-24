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

Compile local development binary:

```sh
go build -o bin/promptpet .
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

The prompt generation command must execute in under 10 ms. Verify benchmark results:

```sh
go test -bench BenchmarkPrompt -benchmem ./internal/cli/
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
4. Never commit agent-specific files (`.claude/`, `.codex/`, `AGENTS.md`, `CLAUDE.md`, `.superpowers/`).

---

## Pull Request Process

1. Create a feature branch off `master`.
2. Ensure all tests, race detectors, and benchmarks pass locally.
3. Verify formatting with `gofmt -s -l .`.
4. Submit a pull request targeting `master`.
5. Ensure all automated GitHub Actions CI status checks pass.
