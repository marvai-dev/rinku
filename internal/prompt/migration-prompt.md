# Before

Check requirements: `rinku req list`
Mark done after implementing: `rinku req done <path>`

# Introduction

You're an experienced senior Go and Rust developer.

You should migrate a Go project to Rust.

The project is in the current directory - **ONLY** read the current directory and children directories
and files in them, **NEVER** it's parents or any other directory outside the current directory.

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
- `<binary>/web/routes/<resource>` - HTTP endpoints
- `<binary>/jobs/<job>` - background jobs
- `db/models/<model>` - database schemas
- `tests/<area>` - test requirements

Example:
  rinku req set api/cli <<EOF
  --port PORT (default: 8080)
  --config FILE (required)
  --verbose
  EOF

Start with step 1.

# Step 1

Analyze the Go project structure. Identify the main entry point and document all packages.

Run `rinku scan go.mod` to see which dependencies have Rust equivalents.

When done, proceed to Step 2.

# Step 2

Create the new project in a IMPORTANT new sub-directory.

When done, proceed to Step 2a.

# Step 2a

Analyze all CLI arguments and options. Capture them as requirements:

  rinku req set <binary>/cli <<EOF
  <flags and options>
  EOF

When done, proceed to Step 2b.

# Step 2b

Analyze all existing tests. Capture them as requirements:

  rinku req set tests/<area> <<EOF
  <test names and descriptions>
  EOF

When done, proceed to Step 10.

# Step 10

Generate the initial Cargo.toml:

Run `rinku convert go.mod -o Cargo.toml`

Review the generated file. Note any unmapped dependencies that need manual research.

When done, proceed to Step 20.

# Step 20

Create the Rust project structure. For each Go package, create a corresponding Rust module:

- `main.go` → `src/main.rs`
- `pkg/foo/foo.go` → `src/foo/mod.rs` or `src/foo.rs`
- `internal/bar/` → `src/bar/` (private module)

When done, proceed to Step 30.

# Step 30
Migrate type definitions. Convert Go structs to Rust structs:

- `type Foo struct` → `struct Foo`
- Pointer fields `*T` → `Option<T>` or `Box<T>`
- Slices `[]T` → `Vec<T>`
- Maps `map[K]V` → `HashMap<K, V>`

When done, proceed to Step 40.

# Step 40
Migrate function signatures. Convert Go functions to Rust:

- `func Foo() error` → `fn foo() -> Result<(), Error>`
- `func Bar(x int) string` → `fn bar(x: i32) -> String`
- Methods `func (f *Foo) Bar()` → `impl Foo { fn bar(&mut self) }`

When done, proceed to Step 50.

# Step 50
Implement error handling. Replace Go error patterns with Rust:

- `if err != nil { return err }` → `?` operator
- Custom errors → implement `std::error::Error` trait
- `panic/recover` → `panic!` / `catch_unwind` (rarely needed)

When done, proceed to Step 55.

# Step 55

Implement CLI based on requirements:

  rinku req get <binary>/cli

After implementing, mark as done:

  rinku req done <binary>/cli

Verify with `rinku req list` - all CLI requirements should show [x].

When all CLI requirements are done, proceed to Step 60.

# Step 60
Run `cargo build` and fix compilation errors. Address each error systematically.

Run `cargo clippy` for additional linting suggestions.

Decide if warnings and errors are migration gaps or due to language differences.
Fix all warnings and errors.

When the project compiles, proceed to Step 65.

# Step 65

Implement tests based on requirements:

  rinku req list tests/
  rinku req get tests/<area>

After implementing each, mark as done:

  rinku req done tests/<area>

Run `cargo test` to verify all tests pass.

If tests fail, decide whether to fix the tests or the code.

Verify all test requirements show [x] in `rinku req list tests/`.

When all test requirements are done, proceed to Step 70.

# Step 70
Add tests. Convert Go tests to Rust:

- `func TestFoo(t *testing.T)` → `#[test] fn test_foo()`
- `t.Errorf()` → `assert!` / `assert_eq!`
- Table-driven tests → use loops or `#[test_case]` macro

Run `cargo test` to verify.

When done, go to steph Finish.

# Step Finish

Create a <project>-migration.md file that describes the migration you did (what, how)
for verification by the user.

Create a README-<project>.md file that describes the migrated project (architecture, structure, frameworks)
for easier onboarding.

When done, migration is complete.

# After

**IMPORTANT:** Mark completed requirements as done:

  rinku req done <path>

Run `rinku req list` to verify all requirements show [x].
