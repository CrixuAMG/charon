# Charon

A TUI project opener for Kitty terminal with Docker support. Named after the ferryman who guides souls across the river Styx, Charon guides you to your projects.

![Charon TUI](https://img.shields.io/badge/TUI-Go-blue)
![License](https://img.shields.io/badge/license-MIT-green)

## Features

- **Project Management** - Add, edit, and delete projects from the TUI
- **Docker Support** - Run commands inside Docker containers
- **Kitty Integration** - Opens projects in new tabs in your current Kitty window
- **Multiple Tasks** - Each project can have multiple tasks, each opening in its own tab
- **Fuzzy Search** - Quickly find projects by typing `/` and searching
- **Filtering** - Filter projects by Docker or Local
- **Sorting** - Sort projects by name, recently used, or custom order
- **Layout Modes** - Switch between card and table layouts (compact and full)
- **Usage Tracking** - Automatically tracks project access in SQLite database
- **Recent Projects** - Sort by recently accessed projects
- **Themes** - Choose from multiple color themes (Catppuccin Mocha, Gruvbox, Tokyo Night)
- **Persistent Config** - Projects are saved to a YAML config file
- **Shell Agnostic** - Works with bash, zsh, nushell, and other shells

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap crixuamg/charon
brew install charon
```

### From Release

Download the latest binary from the [releases page](https://github.com/crixuamg/charon/releases):

```bash
# Linux (amd64)
curl -L https://github.com/crixuamg/charon/releases/latest/download/charon-linux-amd64 -o charon
chmod +x charon
sudo mv charon /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/crixuamg/charon/releases/latest/download/charon-darwin-arm64 -o charon
chmod +x charon
sudo mv charon /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/crixuamg/charon/releases/latest/download/charon-darwin-amd64 -o charon
chmod +x charon
sudo mv charon /usr/local/bin/
```

### From Source

```bash
git clone https://github.com/crixuamg/charon.git
cd charon
go build -o charon .
sudo mv charon /usr/local/bin/
```

## Requirements

- **Kitty Terminal** with remote control enabled
- **Docker** (optional, for Docker-based projects)

### Kitty Configuration

Add the following to your `~/.config/kitty/kitty.conf`:

```
allow_remote_control yes
listen_on unix:/tmp/kitty
```

Restart Kitty after making these changes.

## Configuration

Charon reads its configuration from `~/.config/charon/.charon.yaml`.

### Usage Tracking

Charon automatically tracks project access in a SQLite database located at `~/.config/charon/charon.db`. This enables:
- Sorting by recently accessed projects
- Tracking access counts
- Future analytics features

The database is created automatically on first run and requires no manual setup.

### Example Configuration

```yaml
# Global settings
docker_path: /var/www/html    # Default Docker path for projects
container: my-container       # Docker container name
theme: catppuccin             # Theme: catppuccin (default), gruvbox, tokyonight

# Interface settings
interface:
  layout: table-compact       # Layout: card, card-compact, table, table-compact

# Projects
projects:
  - name: my-app
    path: ~/projects/my-app           # Local path (used for local projects)
    docker_path: /var/www/html        # Docker path (overrides global, empty = local)
    tasks:
      - lazygit
      - vim
      - yarn dev
      - db

  - name: local-project
    path: ~/projects/local-project
    docker_path: ""                   # Empty = local project (no Docker)
    tasks:
      - lazygit
      - vim
```

### Configuration Fields

| Field | Description |
|-------|-------------|
| `docker_path` | Global default path inside Docker container |
| `container` | Docker container name for `docker exec` |
| `theme` | Color theme: `catppuccin` (default), `gruvbox`, `tokyonight` |
| `interface.layout` | Display layout: `card`, `card-compact`, `table`, `table-compact` |
| `projects` | Array of project configurations |

### Project Fields

| Field | Description |
|-------|-------------|
| `name` | Project name (displayed in TUI) |
| `path` | Local filesystem path to the project |
| `docker_path` | Path inside Docker container (empty = local project) |
| `tasks` | Array of commands to run in separate tabs |

## Usage

```bash
charon
```

### Keyboard Controls

#### List View

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` / `Home` | Go to first project |
| `G` / `End` | Go to last project |
| `Enter` / `Space` | Open project (creates tabs) |
| `/` | Search projects (fuzzy search) |
| `s` | Cycle sort mode (Name → Recent → Custom) |
| `f` | Cycle filter (All → Docker → Local) |
| `l` | Cycle layout (Card → Card Compact → Table → Table Compact) |
| `a` | Add new project |
| `e` | Edit selected project |
| `d` / `x` | Delete selected project |
| `q` / `Ctrl+C` | Quit |

#### Search Mode

| Key | Action |
|-----|--------|
| Type characters | Search for projects |
| `Enter` | Accept search and return to list |
| `Esc` | Clear search and return to list |
| `Backspace` | Delete last character (or exit if empty) |

#### Form View (Add/Edit)

| Key | Action |
|-----|--------|
| `Tab` / `↓` | Next field |
| `Shift+Tab` / `↑` | Previous field |
| `Ctrl+S` | Save project |
| `Esc` | Cancel |

#### Delete Confirmation

| Key | Action |
|-----|--------|
| `y` / `Enter` | Confirm delete |
| `n` / `Esc` | Cancel |

### Layout Modes

Charon supports four different layout modes for displaying projects:

#### Card Layout (Default)
- Full project cards with all details
- Shows: name, type badge, path, and all tasks
- Best for: Detailed overview with fewer projects

#### Card Compact Layout
- Simplified cards without task lists
- Shows: name, type badge, and path only
- Best for: Quick scanning with many projects

#### Table Layout
- Tabular view with aligned columns
- Shows: name, type, path, and tasks in columns
- Best for: Comparing projects side-by-side

#### Table Compact Layout
- Minimal table without tasks column
- Shows: name, type, and path in columns
- Best for: Maximum density, viewing many projects at once

Press `l` to cycle through layouts. Your preference is automatically saved to the config file.

## How It Works

### Local Projects

For local projects (empty `docker_path`), Charon:
1. Opens a new Kitty tab with an interactive shell
2. Sends `cd <path>` to navigate to the project
3. Sends the task command (e.g., `lazygit`)

### Docker Projects

For Docker projects, Charon:
1. Opens a new Kitty tab running `docker exec -it <container> $SHELL`
2. Waits for the shell to initialize
3. Sends `cd <docker_path>/<project>` to navigate
4. Sends the task command

### Tab Titles

Each task opens in a new Kitty tab with an appropriate title:

| Command | Tab Title |
|---------|-----------|
| `lazygit` | git |
| `vim` / `nvim` | editor |
| `yarn dev` / `npm run dev` | dev |
| `db` | database |
| Other | First word of command |

## License

MIT License - see [LICENSE](LICENSE) for details.
