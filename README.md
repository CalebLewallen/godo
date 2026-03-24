![godo-gopher](/assets/static/godo-gopher.png)

# GoDo

A keyboard-driven terminal todo app with folder organization, markdown descriptions, and persistent SQLite storage. I wasn't happy with the todo apps I had access to at work, so I made this one. I'll keep refining it, but it's setup for how I like to use TUI apps, which may be weird to you.

## Features

- Hierarchical folder and task organization
- Markdown support in task descriptions
- Fully keyboard-driven with a VS Code-style quick-open task switcher
- Persistent storage via SQLite (`~/.local/share/godo/godo.db`)
- Self-update and uninstall via CLI flags

## Installation

```bash
go install github.com/CalebLewallen/godo@latest
```

Or build from source:

```bash
git clone https://github.com/CalebLewallen/godo
cd godo
go build -ldflags "-X main.version=v0.1.0" -o godo .
```

## Usage

```bash
godo                 # launch the app
godo --version       # print version
godo --update        # update to the latest release
godo --uninstall     # remove the binary and all data
```

## Keyboard Shortcuts

### Global

| Key | Action |
|-----|--------|
| `ctrl+p` | Quick open task |
| `ctrl+e` | Toggle focus sidebar ↔ task |
| `ctrl+b` | Toggle sidebar |
| `ctrl+f` | Filter sidebar |
| `ctrl+n` | New task (in active folder) |
| `alt+ctrl+n` | New folder (prompts for name) |
| `ctrl+q` | Quit |
| `?` | Show help modal |

### Sidebar

| Key | Action |
|-----|--------|
| `↑ / ↓` or `k / j` | Navigate |
| `→ / ←` or `l / h` | Expand / Collapse folder |
| `enter` | Open task / Toggle folder |
| `shift+tab` | Switch tabs (Todo / Completed) |
| `f2` | Rename selected folder or task |
| `shift+→` | Indent folder/task (nest under folder above) |
| `shift+←` | Dedent folder/task (move up one level) |
| `shift+↑ / ↓` | Reorder folder/task among siblings |
| `ctrl+→` | Expand all folders |
| `ctrl+←` | Collapse all folders |
| `ctrl+x` | Delete selected folder or task (prompts to confirm) |

### Task Panel

| Key | Action |
|-----|--------|
| `tab / shift+tab` | Next / Previous field |
| `ctrl+s` | Save task |
| `ctrl+d` | Toggle task done / incomplete |
| `ctrl+w` | Close task (prompts if unsaved) |
| `ctrl+o` | Open all links in description |

## Data

All data is stored in `~/.local/share/godo/godo.db` (SQLite). The database is created automatically on first run. Schema migrations run automatically on startup.

Deleting a folder also deletes all tasks and subfolders it contains.
