# Poli 🏓

**Poli** is a blazing fast, terminal-based HTTP client designed as a lightweight, "techy" alternative to Postman. Built with Go, Poli allows you to manage collections of API requests, import directly from `curl`, and view JSON responses in beautiful, terminal-friendly tables.

No Electron. No lag. Just your terminal.

## Features

- **🚀 Native & Fast:** Written in Go, zero startup lag.
- **📂 Collections:** Organize your endpoints into manageable groups.
- **🔄 Curl Import:** Paste a `curl` command to instantly save a request.
- **📊 Tabular JSON:** JSON responses are automatically rendered as clean, readable tables.
- **🔍 Deep Dive:** Drill into nested objects with `--expand`, view specific array items with `--row`, or filter arrays with `--search`.
- **🕒 Offline History:** Automatically saves your last response so you can view it later without hitting the API again.
- **📝 Interactive Forms:** Create and edit requests with simple CLI prompts.

## Installation

Make sure you have Go installed, then run:

```bash
go install github.com/jojipanackal/poli@latest
```

*Ensure your `~/go/bin` (or `GOPATH/bin`) is in your system's `PATH`.*

## Quick Start

### 1. Create a Group (Collection)
Organize your requests into a group.
```bash
poli new group "JSONPlaceholder"
```

### 2. Import a Request from Curl
The fastest way to add a request is to paste a curl command.
```bash
poli new "Get All Posts" --curl 'curl https://jsonplaceholder.typicode.com/posts'
```

### 3. Ping the Request
Hit the API. JSON array responses are automatically formatted as multi-column tables capping at 10 rows.
```bash
poli ping "Get All Posts"
```

## Usage & Commands

### Managing Requests

*   **List requests in the current group:**
    ```bash
    poli list
    ```
*   **List all groups:**
    ```bash
    poli list --groups
    ```
*   **Switch to a different group:**
    ```bash
    poli use "My Other Group"
    ```
*   **Create a request interactively:**
    ```bash
    poli new "Manual Request"
    ```
*   **Edit an existing request:**
    ```bash
    poli edit "Get All Posts"
    ```
*   **Show request details and underlying curl command:**
    ```bash
    poli show "Get All Posts"
    poli show "Get All Posts" --curl
    ```

### Executing and Viewing Responses

Use `poli ping` to execute a request, or `poli last` to view the previously saved response without making a new network call. Both commands share the same powerful output flags.

*   **Show Headers (Tabular format):**
    ```bash
    poli ping "Get All Posts" --headers
    ```
*   **Show Full Raw Response (Status, Headers, and Raw JSON body):**
    ```bash
    poli ping "Get All Posts" --full --raw
    ```

### Drilling into JSON

Poli's table renderer makes it easy to navigate complex JSON without leaving the terminal.

*   **Expand a nested JSON object key:**
    ```bash
    poli ping "Get User" --expand address
    ```
*   **Expand a specific row in a JSON array:**
    ```bash
    poli last "Get All Posts" --row 12
    ```
*   **Search/Filter an array by value:**
    ```bash
    poli last "Get All Posts" --search "odit"
    ```
*   **Search/Filter an array by `key=value`:**
    ```bash
    poli last "Get All Posts" --search "userId=5"
    ```
*   *(Tip: You can combine `--row` or `--search` with `--raw` to extract the raw JSON of a specific item!)*

## Data Storage
Poli stores your groups and requests locally as plain JSON files in `~/.poli/groups/`. They are highly inspectable, easy to back up, and version-control friendly. Your configuration is stored at `~/.poli/config.yaml`.
