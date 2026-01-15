# rinku

*Disclaimer: This software is mostly written by an AI. I understand this might change your feelings about the software and you might not want to use it.*

**180+ curated Go-to-Rust library mappings** — one of the largest open databases of equivalent libraries for migrating Go projects to Rust.

A CLI tool that instantly finds equivalents for libraries. Give it a GitHub URL, get back the best  alternative in your target language.

Currently works for:
* Go -> Rust

In the works:
* JavaScript -> Go
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

### `migrate` - AI-assisted migration workflow

```bash
rinku migrate
```

A guided multi-step workflow for migrating Go projects to Rust with an AI assistant. Run your agent in your Go project directory and give your AI this prompt:

```
Execute `rinku migrate` and follow instructions.
```

The workflow guides the AI through analyzing the project, creating the Rust structure, converting types and functions, migrating tests and APIs and verifying the migration.

## Coverage

**180+ library mappings** covering 300+ libraries across 25+ categories:

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
