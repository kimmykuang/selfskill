# ss — Skills & Plugin Manager for Claude Code

`ss` is a local CLI tool that manages skills, prompts, and plugins for [Claude Code](https://docs.anthropic.com/en/docs/claude-code). It solves the problem of context bloat by letting you organize skills into groups and load them on demand, rather than having everything active at once.

## Features

- **Skill management** — Install skills from URLs or local directories, organize into groups, load/unload via symlinks
- **Plugin management** — Full takeover of CC's plugin system: add marketplaces, install/load/unload plugins independently
- **Prompt library** — Store and search reusable prompts with tags and metadata
- **Group system** — Bundle skills and plugins together, load entire groups with one command
- **Web UI** — Browser-based management interface for prompts and plugins
- **Single binary** — No runtime dependencies, static assets embedded

## Installation

```bash
# From source
go install github.com/kimmykuang/selfskill/cmd/ss@latest

# Or build locally
git clone https://github.com/kimmykuang/selfskill.git
cd selfskill
make build
# Binary at ./bin/ss
```

## Quick Start

```bash
# Install a skill from a local directory
ss install skill ./my-skills/video-summary/

# Create a group and add skills
ss group create work
ss group add work video-summary

# Load the group (creates symlinks in ~/.claude/skills/)
ss load work

# Check what's active
ss status

# Unload when done
ss unload work
```

## Plugin Management

```bash
# Add a marketplace (clone its repo to get plugin metadata)
ss plugin add-marketplace https://github.com/anthropics/claude-plugins-official.git

# See what's available
ss plugin list-available

# Install a plugin (downloads to ~/ss/plugins/, does NOT load)
ss plugin install superpowers@claude-plugins-official

# Load into CC (creates symlink + registers in installed_plugins.json)
ss plugin load superpowers

# Unload from CC
ss plugin unload superpowers

# Add plugin to a group
ss group add --plugin work superpowers
ss load work   # loads both skills and plugins
```

## Commands

| Command | Description |
|---------|-------------|
| `ss install skill <source>` | Install skill from URL or local path |
| `ss install prompt <source>` | Install prompt from URL or local path |
| `ss skill list` | List installed skills |
| `ss skill remove <name>` | Remove a skill |
| `ss prompt list` | List installed prompts |
| `ss prompt remove <id>` | Remove a prompt |
| `ss group create <name>` | Create a skill group |
| `ss group add <group> <skill>` | Add skill to group |
| `ss group add --plugin <group> <plugin>` | Add plugin to group |
| `ss group remove <group> <item>` | Remove item from group |
| `ss group list` | List all groups |
| `ss group delete <name>` | Delete a group |
| `ss load <group>` | Load group (skills + plugins) |
| `ss unload <group>` | Unload group |
| `ss unload --all` | Unload all ss-managed skills |
| `ss status` | Show active skills and plugins |
| `ss plugin add-marketplace <url>` | Add a plugin marketplace |
| `ss plugin list-available` | List installable plugins |
| `ss plugin install <plugin>@<marketplace>` | Download a plugin |
| `ss plugin load <plugin>` | Activate plugin in CC |
| `ss plugin unload <plugin>` | Deactivate plugin from CC |
| `ss plugin list` | List ss-managed plugins |
| `ss plugin skills <plugin>` | List skills in a plugin |
| `ss plugin update <plugin>` | Git pull latest version |
| `ss plugin remove <plugin>` | Delete plugin source |
| `ss web [--port 8484]` | Start web management UI |

## Directory Structure

```
~/ss/
├── skills/           # Installed skill directories
│   └── video-summary/
│       ├── SKILL.md
│       ├── assets/
│       └── scripts/
├── prompts/          # Prompt markdown files
├── groups/           # Group YAML configs
├── plugins/          # Plugin source (git clones)
│   └── claude-plugins-official/
│       └── superpowers/
│           └── 5.1.0/
├── marketplaces/     # Marketplace repos (for plugin discovery)
└── ss.yaml           # Global config
```

## How It Works

**Skills** are loaded via symlinks:
```
~/.claude/skills/video-summary -> ~/ss/skills/video-summary/
```

**Plugins** are loaded via symlinks + JSON registration:
```
~/.claude/plugins/cache/mp/plugin/ver -> ~/ss/plugins/mp/plugin/ver/
+ entry in ~/.claude/plugins/installed_plugins.json
```

CC reads these paths on startup — no restart needed for skills, restart needed for plugins.

## Web UI

```bash
ss web --port 8484
# Open http://localhost:8484
```

Pages: Prompts (CRUD + search), Skills (read-only), Plugins (load/unload), Status.

## Development

```bash
make build    # Build binary
make test     # Run tests with race detector
make clean    # Remove build artifacts
```

## License

MIT
