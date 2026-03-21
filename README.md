# poli

![GitHub Release](https://img.shields.io/github/v/release/jojipanackal/poli?style=flat-square)
![Go Report Card](https://goreportcard.com/badge/github.com/jojipanackal/poli?style=flat-square)
![License](https://img.shields.io/github/license/jojipanackal/poli?style=flat-square)

```text
    ____        ___
   / __ \____  / (_)
  / /_/ / __ \/ / / 
 / ____/ /_/ / / /  
/_/    \____/_/_/   
```

A terminal-based HTTP client for managing and executing collections of API requests. Designed for speed and minimal latency, `poli` supports request organization, `curl` importation, and structured JSON response rendering.

## Core Features

- **Performance**: Native Go binary with zero startup overhead.
- **Collection Management**: Organize requests into logical groups.
- **CURL Integration**: Import requests directly from `curl` commands.
- **Data Rendering**: Automatic tabular formatting for JSON responses.
- **Offline Access**: Persists last-received responses for offline inspection.
- **Index Support**: Reference requests as `r1`, `r2` and groups as `g1`, `g2` for faster navigation.
- **Shell Completion**: Supports Bash, Zsh, and Fish.

## Installation

### Homebrew (macOS/Linux)
The recommended way to install `poli` is via Homebrew:

```bash
brew tap jojipanackal/tap
brew install poli
```

### Go Install
If you have Go installed on your system:

```bash
go install github.com/jojipanackal/poli@latest
```

## Indexing System

`poli` allows you to reference requests and collections by their index numbers shown in the `list` command. This saves you from typing out long, descriptive names.

- **Requests**: Use `r1`, `r2`, etc. to reference requests in the active group.
- **Groups**: Use `g1`, `g2`, etc. to reference collections.

Example: `poli ping r1` instead of `poli ping "Get All Users From Production"`

## Help View

```text
Terminal-based HTTP client — collections, requests, curl import, zero lag.

Usage:
  poli [command]

Request Operations:
  edit        Edit a saved request
  list        List requests in the current group
  new         Create a new request or group
  ping        Execute a saved request
  use         Switch to a different group/collection

Collection Management:
  delete      Delete a request or group
  last        Show the last response for a request
  mv          Move a request to another group
  show        Show details of a saved request

Utility Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print the version number of poli

Flags:
      --config string   config file (default is $HOME/.poli/config.yaml)
  -g, --group string    active group/collection (default from config)
  -h, --help            help for poli
```

## Usage

### Creating a Collection
```bash
poli new group "API-V1"
```

### Importing a Request
```bash
poli new "Get-Users" --curl 'curl https://api.example.com/users'
```

### Executing Requests
```bash
poli ping "Get-Users"
# Or using index
poli ping r1
```

### Inspecting Responses
- `--headers`: Show response headers.
- `--expand <key>`: Drill into nested JSON objects.
- `--row <n>`: View a specific row in a JSON array.
- `--search <query>`: Filter array results by value or `key=value`.

### Navigation
- `poli use g1`: Switch to the first collection.
- `poli show r2`: View details of the second request in current group.

## Storage

Configuration and data are persisted locally:
- **Config**: `~/.poli/config.yaml`
- **Data**: `~/.poli/groups/` (JSON format)
