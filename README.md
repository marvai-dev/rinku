# rinku

*Disclaimer: This software is mostly written by an AI. I understand this might change your feelings about the software and not use it.*

**160+ curated Go-to-Rust library mappings** — one of the largest open databases of equivalent libraries for migrating Go projects to Rust.

A CLI tool that instantly finds equivalents for libraries. Give it a GitHub URL, get back the best  alternative in your target language.

Currently works for:
* Go -> Rust
* JavaScript -> Go

In the works:
* Python -> Rust

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap marvai-dev/rinku
brew install rinku
```

### Download Binary

Download from [GitHub Releases](https://github.com/marvai-dev/rinku/releases), or:

```bash
# Linux (amd64)
curl -sL https://github.com/marvai-dev/rinku/releases/latest/download/rinku_linux_amd64.tar.gz | tar xz
sudo mv rinku /usr/local/bin/

# macOS (Apple Silicon)
curl -sL https://github.com/marvai-dev/rinku/releases/latest/download/rinku_darwin_arm64.tar.gz | tar xz
sudo mv rinku /usr/local/bin/
```

### With Go

```bash
go install github.com/marvai-dev/rinku/cmd/rinku@latest
```

### From Source

```bash
make build
```

## Usage

### `lookup` - Find equivalent library

```bash
rinku lookup <url> [language]
```

Look up an equivalent library for a GitHub URL. Defaults to Rust target.

```bash
# Go → Rust (default)
rinku lookup https://github.com/spf13/cobra
# Output: https://github.com/clap-rs/clap

# JavaScript → Go
rinku lookup https://github.com/lodash/lodash go
# Output: https://github.com/samber/lo

# Include libraries with known vulnerabilities
rinku lookup https://github.com/golang/net --unsafe
```

### `scan` - Analyze go.mod

```bash
rinku scan <path>
```

Parse a go.mod file and show equivalents for each dependency.

```bash
rinku scan ./go.mod
```

### `convert` - Generate Cargo.toml

```bash
rinku convert <path>
```

Generate a Cargo.toml from a go.mod file.

```bash
rinku convert ./go.mod > Cargo.toml
```

## Security

Rinku excludes libraries with known security vulnerabilities by default. Use `--unsafe` to include them.

To validate all library mappings for security issues:

```bash
make validate
```

This checks all Rust library targets for:
- Repository existence and health (via GitHub API)
- Known vulnerabilities (via OSV API)

## Coverage

**160+ library mappings** covering 270+ libraries across 25+ categories:

| Category | Examples |
|----------|----------|
| Web Frameworks | gin → axum, echo → axum |
| CLI | cobra → clap, viper → config-rs |
| Serialization | yaml → serde-yaml, json → serde-json, protobuf → prost |
| Observability | opentelemetry → opentelemetry-rust, prometheus → client_rust |
| Logging | zap → tracing, logrus → tracing, zerolog → tracing |
| HTTP/gRPC | grpc-go → tonic, net/http → hyper |
| Database | gorm → sea-orm, sqlx → sqlx |
| Async/Concurrency | goroutines → tokio, channels → crossbeam |
| ...and more | crypto, compression, kubernetes, docker, etc. |

## License

FSL-1.1-MIT
