# Rinku Internals

## What is Rinku?

Rinku is a CLI tool for migrating Go projects to Rust with AI assistance. It provides:

1. **Library Mapping**: Database of 180+ Go-to-Rust library equivalents across 25+ categories
2. **Guided Migration**: Multi-step workflow with progress tracking and requirement capture

## Request Set (Requirements System)

The requirement system captures specifications during migration. Requirements are hierarchical specs that track what needs to be implemented.

### Storage

Requirements are stored as individual JSON files in `.rinku/progress/requirements/`:
```
.rinku/progress/requirements/
├── api/
│   └── users.json
├── cli.json
└── db/
    └── models.json
```

### Data Model

Each requirement contains:
- **Path**: Hierarchical identifier (e.g., `api/users`, `cli`)
- **Content**: The specification text
- **Step**: Which migration step created it
- **Done**: Completion status
- **Timestamps**: Created/Updated/Done timestamps

### Commands

```bash
rinku req set <path> <content>   # Create/update requirement
rinku req get <path>             # View requirement
rinku req list [prefix]          # List all requirements
rinku req done <path>            # Mark as completed
```

### Usage

```bash
# Capture CLI flags
rinku req set myapp/cli << EOF
--port: Port to listen on (default: 8080)
--verbose: Enable verbose logging
EOF

# Check progress
rinku req list
# [x] myapp/cli
# [ ] myapp/api/users
```

## Steps (Migration Workflow)

The migration workflow is a guided, multi-step process defined in `migration-prompt.md`.

### Step States

```
pending → in_progress → completed
                     ↘ skipped
```

### Progress Tracking

Progress is stored in `.rinku/progress.json`:
```json
{
  "version": 1,
  "current_step": "step-2",
  "steps": {
    "step-1": {"status": "completed", "started_at": "...", "finished_at": "..."},
    "step-2": {"status": "in_progress", "started_at": "..."}
  },
  "step_order": ["step-1", "step-2", "step-3"]
}
```

### Commands

```bash
rinku migrate                    # Show introduction
rinku migrate <step-id>          # Show step content
rinku migrate --start <step>     # Mark step as in_progress
rinku migrate --finish <step>    # Mark step as completed
rinku migrate --status           # Show progress summary
rinku migrate --reset            # Clear progress and restart
```

### Workflow

1. AI shows introduction with `rinku migrate`
2. Starts step with `rinku migrate --start step-1`
3. Captures requirements with `rinku req set ...`
4. Completes step with `rinku migrate --finish step-1`
5. Repeats for each step

## Key Packages

| Package | Purpose |
|---------|---------|
| `progress` | Migration step tracking and persistence |
| `requirements` | Requirement storage with path validation |
| `multistep` | Parses markdown prompts into steps |
| `prompt` | Embeds and loads migration-prompt.md |
| `rinku` | Library mapping database and lookup |
| `gomod` | Parses go.mod for dependencies |
| `cargo` | Generates Cargo.toml from mappings |
| `types` | Shared data structures (Library, Mapping) |

## Storage Layout

```
.rinku/
├── progress.json                    # Step progress tracking
└── progress/
    └── requirements/                # Requirement JSON files
        └── <path>.json
```

All storage uses atomic writes to prevent corruption. The `SafeReqPath` type prevents directory traversal attacks in requirement paths.
