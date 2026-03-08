# poli

A terminal-based HTTP client for managing and executing collections of API requests. Designed for speed and minimal latency, `poli` supports request organization, `curl` importation, and structured JSON response rendering.

## Core Features

- **Performance**: Native Go binary with zero startup overhead.
- **Collection Management**: Organize requests into logical groups.
- **CURL Integration**: Import requests directly from `curl` commands.
- **Data Rendering**: Automatic tabular formatting for JSON responses.
- **Offline Access**: Persists last-received responses for offline inspection.

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
  version     Print the version number of poli

Additional Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command

Flags:
      --config string   config file (default is $HOME/.poli/config.yaml)
  -g, --group string    active group/collection (default from config)
  -h, --help            help for poli

Use "poli [command] --help" for more information about a command.
```

## Installation

```bash
go install github.com/jojipanackal/poli@latest
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
```

### Inspecting Responses
- `--headers`: Show response headers.
- `--expand <key>`: Drill into nested JSON objects.
- `--row <n>`: View a specific row in a JSON array.
- `--search <query>`: Filer array results by value or `key=value`.

## Storage

Configuration and data are persisted locally:
- **Config**: `~/.poli/config.yaml`
- **Data**: `~/.poli/groups/` (JSON format)
