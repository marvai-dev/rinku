# Before

**FIRST:** Run `rinku req list` before **EACH** step to see pending requirements you must implement. **ONLY** implement requirements that are relevant to the current step.

# After

**IMPORTANT:** Mark completed requirements as done:

  rinku req done <path>

Run `rinku req list` to verify all requirements that need to be done in this step show [x].

# Introduction

You're an experienced senior Go and Rust developer.

You should migrate a Go project to Rust.

The project is in the current directory - **ONLY** read the current directory and children directories
and files in them, **NEVER** its parents or any other directory outside the current directory.

This is a multi-step process.

Use todos and a plan to guide your work.

- Use `rinku migrate --start <step>` to start a step (shows instructions).
- Use `rinku migrate --finish <step>` to mark a step as complete.
- Use `rinku migrate --status` to see progress.
- Use `rinku migrate --reset` to start over.

The surface of the application (APIs) needs to be the same in Rust as in Go.
Use requirements to track what must work in Rust:

  rinku req set <path> <<EOF         # capture a requirement (use heredoc for multi-line)
  <content>
  EOF
  rinku req get <path>               # view a requirement
  rinku req list [prefix]            # list with status [x] done, [ ] pending
  rinku req done <path>              # mark as implemented

Suggested paths:
- `<binary>/cli` - command line flags
- `<binary>/api/<resource>` - HTTP endpoints
- `<binary>/templates/<name>` - web templates
- `<binary>/static` - static asset configuration
- `<binary>/middleware` - middleware stack
- `<binary>/sessions` - session management
- `<binary>/jobs/<job>` - background jobs
- `db/models/<model>` - database schemas
- `tests/<area>` - test requirements

Example:
  rinku req set myapp/cli <<EOF
  --port PORT (default: 8080)
  --config FILE (required)
  --verbose
  EOF

Start with Step 1.

# Step 1

Analyze the Go project structure. Identify the main entry point and document all packages.

Run `rinku scan go.mod` to see which dependencies have Rust equivalents.

Run `rinku analyze go.mod` to detect project type. The output shows which features are used:
- `cli` - has CLI framework (Steps 3, 16 relevant)
- `web` - has web framework (Steps 4-8, 17-21 relevant)
- `sql`, `orm` - has database layer
- `grpc` - has gRPC

Skip steps for features not detected (e.g., if no `web` tag, skip Steps 5-8 and 18-21).

When done, proceed to Step 2.

# Step 2

Create the Rust project in a new sub-directory:

```bash
cargo new <project-name>
cd <project-name>
```

When done, proceed to Step 3.

# Step 3

Capture CLI arguments and options as requirements:

  rinku req set <binary>/cli <<EOF
  <flags and options>
  EOF

When done, proceed to Step 4.

# Step 4

Capture HTTP/API endpoints as requirements (skip if no web server):

  rinku req set <binary>/api/<resource> <<EOF
  GET /users - list users
  POST /users - create user
  ...
  EOF

When done, proceed to Step 5.

# Step 5

Capture web templates as requirements (skip if no templates):

  rinku req set <binary>/templates/<name> <<EOF
  Source: templates/user.html
  Variables: .Name, .Email, .Items[]
  Layout: extends base.html
  EOF

When done, proceed to Step 6.

# Step 6

Capture static asset configuration (skip if no static files):

  rinku req set <binary>/static <<EOF
  Directory: static/
  Mount path: /static
  Files: CSS, JS, images
  EOF

When done, proceed to Step 7.

# Step 7

Capture middleware requirements (skip if no middleware):

  rinku req set <binary>/middleware <<EOF
  - Auth: JWT validation on /api/* routes
  - CORS: Allow origins X, Y
  - Logging: Request/response logging
  EOF

When done, proceed to Step 8.

# Step 8

Capture session requirements (skip if no session management):

  rinku req set <binary>/sessions <<EOF
  Store: Redis / Memory / Cookie
  Cookie name: session_id
  Expiry: 24h
  EOF

When done, proceed to Step 9.

# Step 9

Capture existing tests as requirements:

  rinku req set tests/<area> <<EOF
  <test names and descriptions>
  EOF

When done, proceed to Step 10.

# Step 10

Generate the initial Cargo.toml in <project-name>/Cargo.toml:

Run `rinku convert go.mod -o Cargo.toml`

Review the generated file. Note any unmapped dependencies that need manual research.

When done, proceed to Step 11.

# Step 11

Create the Rust project structure. For each Go package, create a corresponding Rust module:

- `main.go` → `src/main.rs`
- `pkg/foo/foo.go` → `src/foo/mod.rs` or `src/foo.rs`
- `internal/bar/` → `src/bar/` (private module)

When done, proceed to Step 12.

# Step 12

Migrate type definitions. Convert Go structs to Rust structs:

- `type Foo struct` → `struct Foo`
- Pointer fields `*T` → `Option<T>` or `Box<T>`
- Slices `[]T` → `Vec<T>`
- Maps `map[K]V` → `HashMap<K, V>`

When done, proceed to Step 13.

# Step 13

Migrate function signatures. Convert Go functions to Rust:

- `func Foo() error` → `fn foo() -> Result<(), Error>`
- `func Bar(x int) string` → `fn bar(x: i32) -> String`
- Methods `func (f *Foo) Bar()` → `impl Foo { fn bar(&mut self) }`

When done, proceed to Step 14.

# Step 14

Implement error handling. Replace Go error patterns with Rust:

- `if err != nil { return err }` → `?` operator
- Custom errors → implement `std::error::Error` trait
- `panic/recover` → `panic!` / `catch_unwind` (rarely needed)

When done, proceed to Step 15.

# Step 15

Migrate concurrency patterns. Convert Go concurrency to Rust:

- `go func()` → `tokio::spawn(async {})` or `std::thread::spawn()`
- `chan T` → `tokio::sync::mpsc` or `std::sync::mpsc`
- `sync.Mutex` → `std::sync::Mutex` or `tokio::sync::Mutex`
- `sync.WaitGroup` → `tokio::join!` or thread handles with `.join()`
- `select {}` → `tokio::select!`

If the project uses goroutines, add `tokio` to Cargo.toml:
```toml
tokio = { version = "1", features = ["full"] }
```

Skip if no concurrency in the Go project.

When done, proceed to Step 16.

# Step 16

Implement CLI based on requirements:

  rinku req get <binary>/cli

After implementing, mark as done:

  rinku req done <binary>/cli

When done, proceed to Step 17.

# Step 17

Implement API endpoints based on requirements (skip if no web server):

  rinku req list <binary>/api/
  rinku req get <binary>/api/<resource>

After implementing each, mark as done:

  rinku req done <binary>/api/<resource>

When done, proceed to Step 18.

# Step 18

Implement web templates based on requirements (skip if no templates):

- Go `html/template` → Rust `askama`, `tera`, or `minijinja`
- Go `text/template` → Rust `tera` or `minijinja`

  rinku req list <binary>/templates/

For each template:
1. Create the equivalent Rust template file
2. Retain the original layout and structure even with a different templating engine
3. Verify the rendered output matches the original
4. Update the handler to use the new template
5. Mark as done: `rinku req done <binary>/templates/<name>`

When done, proceed to Step 19.

# Step 19

Implement static file serving based on requirements (skip if no static files):

- Go `http.FileServer` → Rust `tower-http::services::ServeDir`

  rinku req get <binary>/static

Implement static file serving and mark as done:

  rinku req done <binary>/static

When done, proceed to Step 20.

# Step 20

Implement middleware based on requirements (skip if no middleware):

- Auth: Go middleware → Rust `tower` layers or framework extractors
- CORS: `rs/cors` → `tower-http::cors::CorsLayer`
- Logging: Go middleware → `tower-http::trace::TraceLayer`
- Rate limiting: Go middleware → `tower::limit` or `governor`

  rinku req get <binary>/middleware

Implement each middleware layer and mark as done:

  rinku req done <binary>/middleware

When done, proceed to Step 21.

# Step 21

Implement session handling based on requirements (skip if no session management):

- `gorilla/sessions` → `tower-sessions`
- Cookie-based auth → `axum-extra::extract::CookieJar` or `actix-web::cookie`

  rinku req get <binary>/sessions

Implement session handling and mark as done:

  rinku req done <binary>/sessions

When done, proceed to Step 22.

# Step 22

Build and verify. Run these commands and fix ALL errors and warnings:

```bash
cargo build
cargo clippy -- -D warnings
```

Do NOT proceed until both commands pass with zero errors and zero warnings.

When the project compiles cleanly, proceed to Step 23.

# Step 23

Implement and run tests based on requirements:

  rinku req list tests/
  rinku req get tests/<area>

Convert Go tests to Rust:
- `func TestFoo(t *testing.T)` → `#[test] fn test_foo()`
- `t.Errorf()` → `assert!` / `assert_eq!`
- Table-driven tests → use loops or `#[test_case]` macro

After implementing each, mark as done:

  rinku req done tests/<area>

Run `cargo test` to verify all tests pass.

When all test requirements are done, proceed to Step 24.

# Step 24

Review the migrated code for:
- Wrongly translated idioms
- Error handling issues
- Missing edge cases
- Non-idiomatic Rust patterns

Fix all problems found.

When done, proceed to Step Finish.

# Step Finish

Create a <project-name>/<project>-migration.md file that describes the migration you did (what, how)
for verification by the user.

Create a <project-name>/README-<project>.md file that describes the migrated project (architecture, structure, frameworks)
for easier onboarding.

When done, migration is complete.
