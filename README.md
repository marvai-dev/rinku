# rinku

![Rinku Mascot](rinku-mascot.webp)

**160+ curated Go-to-Rust library mappings** — one of the largest open databases of equivalent libraries for migrating Go projects to Rust.

A CLI tool that instantly finds Rust equivalents for Go libraries. Give it a GitHub URL, get back the best Rust alternative.

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

```bash
rinku <github-url> [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--unsafe` | Include libraries with known vulnerabilities in results |
| `-h, --help` | Show help |

### Examples

Find the Rust equivalent of a Go CLI framework:

```bash
rinku https://github.com/spf13/cobra
# Output: https://github.com/clap-rs/clap
```

Find the Rust equivalent of a Go web framework:

```bash
rinku https://github.com/gin-gonic/gin
# Output: https://github.com/tokio-rs/axum
```

Include libraries with known vulnerabilities:

```bash
rinku https://github.com/golang/net --unsafe
# Output: https://github.com/hyperium/hyper
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
