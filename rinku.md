# Plan: Enhance Rinku's Migration Workflow with Iteration

## Goal

Enhance the existing 24-step prompt sequence with:
1. **Iteration loops** - each step loops until stable (no more findings)
2. **Gated progression** - can only proceed when current step's requirements are complete
3. **Codegen support** - dedicated step for code generation
4. **Lightweight verification** - `rinku verify` shows expected vs captured

**Key insight:** The AI understands code. Don't tell it *what* to look for. Tell it *how to iterate* and *how to verify completeness*.

**Non-interactive:** `rinku migrate --start N` outputs instructions, AI executes autonomously.

---

## The Problems

1. **AI is probabilistic** - might miss things on first pass
   - Solution: Loop until a full pass finds nothing new

2. **Context window limits** - can't load entire codebase
   - Solution: Steps break up work by concern (CLI, API, templates, etc.)

3. **No verification** - AI doesn't know when it's done
   - Solution: `rinku verify` shows expected categories vs what's captured

---

## Proposed Changes

### 1. `rinku verify` - Expected vs Captured

Based on detected tags, show what categories are expected and what's been captured:

```bash
$ rinku verify go.mod
```

```
Expected (from tags): cli, web, sql, codegen:protobuf

Captured:
  [x] cli - 1 requirement
  [ ] web - 0 requirements (EXPECTED but missing)
  [ ] sql - 0 requirements (EXPECTED but missing)
  [ ] codegen:protobuf - 0 requirements (EXPECTED but missing)
  [x] tests - 2 requirements (not expected, but captured - OK)

Status: 2/4 expected categories covered
```

Check implementation status:

```bash
$ rinku verify --impl
```

```
Requirements: 8/12 implemented
  [ ] myapp/api/auth
  [ ] myapp/api/posts
```

**Note:** `verify` only checks requirements. The AI runs `cargo build`, `cargo test`, etc.

### 2. No `rinku guide` - Trust the AI

~~`rinku guide`~~ - Removed. The AI is better at understanding code than rigid guidance.

Instead, the **prompt** tells the AI:
- How to document findings (requirement format)
- How to iterate (loop until stable)
- How to verify (run `rinku verify`)

### 3. Iteration Pattern: Loop Until Stable

The AI is probabilistic - it might miss things. Solution: **loop until a full pass finds nothing new**.

**Capture step pattern:**
```markdown
# Step 3: Capture CLI

Find and document all CLI requirements.

Iteration (max 5 passes):
1. Read relevant source files (main.go, cmd/*.go, etc.)
2. For each CLI feature found, record it:
   rinku req set <binary>/cli/<feature> <<EOF
   <description>
   EOF
3. Do another pass through the code
4. If you found new features, go to step 2
5. When a full pass finds nothing new, OR you've done 5 passes, proceed

The goal: capture EVERYTHING. Multiple passes catch what you missed.
```

**Implement step pattern:**
```markdown
# Step 17: Implement API

Implement all API requirements.

Iteration:
1. Run `rinku req list <binary>/api/`
2. For each pending requirement:
   - `rinku req get <path>` to see details
   - Implement in Rust
   - `rinku req done <path>`
3. Run `rinku verify --impl`
4. If pending API requirements remain, go to step 1
5. When all done, proceed to Step 18

(No limit - must complete all requirements before proceeding)
```

### 4. Gated Progression

Steps can only complete when their requirements are done.

**Enforcement:** `rinku migrate --finish N` checks that requirements for step N are complete. Fails if not.

### 5. Dedicated Codegen Step

New step after Step 9 (after all requirements captured, before Cargo.toml generation):

```markdown
# Step Codegen

Capture code generation requirements.

Look for:
- `//go:generate` directives in source files
- *.proto files (protobuf)
- ent/schema/*.go (ent ORM)
- *.templ files (templ)
- wire.go files (wire DI)
- sqlc.yaml (sqlc)
- gqlgen.yml (gqlgen)

For each codegen source found:
  rinku req set codegen/<tool> <<EOF
  Source files: <paths>
  Generated files: <paths>
  Purpose: <what it generates>
  EOF

Rust equivalents (for reference when implementing):
- protobuf → prost + tonic-build (build.rs)
- ent → sea-orm or diesel
- templ → askama
- wire → manual DI or shaku
- sqlc → sqlx macros or sea-query
- gqlgen → async-graphql

Iteration: Loop until a full pass finds no new codegen sources.
When done, proceed to Step 10.
```

### 6. New Tags for Codegen

Add to `libs.json`:

```json
"go:google.golang.org/protobuf": { "tags": ["codegen:protobuf"] },
"go:entgo.io/ent": { "tags": ["orm", "codegen:ent"] },
"go:github.com/a-h/templ": { "tags": ["templating", "codegen:templ"] },
"go:github.com/google/wire": { "tags": ["di", "codegen:wire"] },
"go:github.com/sqlc-dev/sqlc": { "tags": ["sql", "codegen:sqlc"] },
"go:github.com/99designs/gqlgen": { "tags": ["graphql", "codegen:gqlgen"] }
```

---

## Updated Step Structure

| Step | Purpose | Requirements Path | Iteration |
|------|---------|-------------------|-----------|
| 1 | Analyze project | - | No |
| 2 | Create Rust project | - | No |
| 3 | Capture CLI | `<binary>/cli/*` | Loop until stable |
| 4 | Capture API | `<binary>/api/*` | Loop until stable |
| 5 | Capture templates | `<binary>/templates/*` | Loop until stable |
| 6 | Capture static | `<binary>/static/*` | Loop until stable |
| 7 | Capture middleware | `<binary>/middleware/*` | Loop until stable |
| 8 | Capture sessions | `<binary>/sessions/*` | Loop until stable |
| 9 | Capture tests | `tests/*` | Loop until stable |
| **Codegen** | **Capture codegen** | `codegen/*` | **Loop until stable** |
| 10 | Generate Cargo.toml | - | No |
| 11 | Create module structure | - | No |
| 12 | Migrate types | - | No |
| 13 | Migrate functions | - | No |
| 14 | Error handling | - | No |
| 15 | Concurrency | - | No |
| 16 | Implement CLI | `<binary>/cli/*` | Loop until all done |
| 17 | Implement API | `<binary>/api/*` | Loop until all done |
| 18 | Implement templates | `<binary>/templates/*` | Loop until all done |
| 19 | Implement static | `<binary>/static/*` | Loop until all done |
| 20 | Implement middleware | `<binary>/middleware/*` | Loop until all done |
| 21 | Implement sessions | `<binary>/sessions/*` | Loop until all done |
| 22 | Build & verify | - | Loop until clean |
| 23 | Implement tests | `tests/*` | Loop until all done |
| 24 | Review | - | No |
| Finish | Documentation | - | No |

---

## New Command

| Command | Purpose |
|---------|---------|
| `rinku verify [go.mod]` | Show expected categories (from tags) vs captured requirements |
| `rinku verify --impl` | Check requirement implementation status |

No `rinku guide` - the AI figures out what to look for.

---

## Implementation Plan

### Phase 1: Verify Command
1. Add `rinku verify` command
   - Map tags to expected requirement categories
   - Compare against captured requirements
   - `--impl` flag checks done status

### Phase 2: Codegen Support
2. Add `codegen:*` tags to `libs.json`
3. Map codegen tags to expected `codegen/*` requirements

### Phase 3: Update Workflow
4. Add "Step Codegen" to `migration-prompt.md`
5. Add "loop until stable" pattern to capture steps (3-9, Codegen)
6. Add "loop until all done" pattern to implement steps (16-21, 23)
7. Add gating to `rinku migrate --finish`

---

## Files to Modify

**Modify:**
- `cmd/rinku/main.go` - Add VerifyCmd
- `cmd/rinku/libs.json` - Add codegen:* tags
- `internal/prompt/migration-prompt.md` - Add codegen step, iteration patterns, gating

**New:**
- `internal/verify/verify.go` - Verify command logic

---

## Example Flow

```bash
# AI starts migration
$ rinku migrate
# Shows introduction with iteration guidance

$ rinku migrate --start 3
# Step 3: Capture CLI - "loop until stable"

# AI reads code, first pass
$ rinku req set myapp/cli/port <<EOF
--port PORT (default: 8080)
EOF

# AI does another pass, finds more
$ rinku req set myapp/cli/config <<EOF
--config FILE (required)
EOF

# AI does another pass, finds nothing new
$ rinku verify go.mod
# cli: 2 requirements (covered)

$ rinku migrate --finish 3
# Gate passes, step complete

# ... later, implementing ...

$ rinku migrate --start 17
# Step 17: Implement API - "loop until all done"

$ rinku req list myapp/api/
# [ ] myapp/api/users
# [ ] myapp/api/auth

# AI implements users endpoint
$ rinku req done myapp/api/users

# AI implements auth endpoint
$ rinku req done myapp/api/auth

$ rinku verify --impl
# All api requirements done!

$ rinku migrate --finish 17
# Gate passes
```

---

## Key Principles

1. **AI does the work** - rinku provides structure and verification only
2. **Loop until stable** - multiple passes catch what was missed
3. **Gated progression** - can't proceed with incomplete requirements
4. **No auto-skip** - AI follows all steps, decides relevance itself
5. **No rigid guidance** - AI understands code better than we can prescribe
