# rinku

![Rinku Mascot](rinku-mascot.webp)

**130+ curated Go-to-Rust library mappings** — one of the largest open database of equivalent libraries for migrating Go projects to Rust.

A CLI tool that instantly finds Rust equivalents for Go libraries. Give it a GitHub URL, get back the best Rust alternative.

## Installation

```shell
go install github.com/marvai-dev/rinku@latest
```

Or build from source:

```bash
make build
```

## Usage

```bash
rinku <github-url> <target-language> [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--unsafe` | Include libraries with known vulnerabilities in results |
| `-h, --help` | Show help |

### Examples

Find the Rust equivalent of a Go CLI framework:

```bash
rinku https://github.com/spf13/cobra rust
# Output: https://github.com/clap-rs/clap
```

Find the Rust equivalent of a Go web framework:

```bash
rinku https://github.com/gin-gonic/gin rust
# Output: https://github.com/tokio-rs/axum
```

Include libraries with known vulnerabilities:

```bash
rinku https://github.com/golang/net rust --unsafe
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

**130+ library mappings** across 25+ categories:

| Category | Examples |
|----------|----------|
| Web Frameworks | gin → axum, echo → actix-web |
| CLI | cobra → clap, viper → config-rs |
| Database | gorm → diesel, sqlx → sqlx |
| Serialization | json → serde, protobuf → prost |
| HTTP Clients | net/http → reqwest, resty → ureq |
| Async/Concurrency | goroutines → tokio, channels → crossbeam |
| Testing | testify → assert, gomock → mockall |
| Logging | zap → tracing, logrus → log |
| ...and more | crypto, compression, encoding, validation, etc. |

## License

FSL-1.1-MIT
