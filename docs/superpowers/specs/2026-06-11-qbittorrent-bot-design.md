# qBittorrent Telegram Bot — Design

## Overview

A Go-based Telegram bot that controls a qBittorrent instance. Users send magnet links or torrent files via Telegram; the bot forwards them to qBittorrent for download, and can query/manage active tasks.

## Architecture

```
┌──────────┐   Telegram Bot API    ┌──────────┐   qBittorrent Web API   ┌──────────────┐
│ Telegram │ ◄── long polling ───► │  Go Bot  │ ─────── HTTP ────────► │  qBittorrent │
│  Client  │                       │ (Docker) │                        │  (anywhere)  │
└──────────┘                       └──────────┘                        └──────────────┘
```

- Telegram communication via `go-telegram-bot-api/telegram-bot-api v5.5.1`
- qBittorrent communication via its Web API (HTTP JSON, using Go stdlib)
- Long polling mode — no webhook URL needed

## Command Set

| Command | Arguments | Description |
|---------|-----------|-------------|
| `/add` | `<magnet URI or URL>` | Add a torrent download |
| (file) | `.torrent` file attachment | Add torrent from file |
| `/list` | — | List all torrents with ID, name, progress%, size, status |
| `/pause` | `<id>` | Pause a torrent |
| `/resume` | `<id>` | Resume a torrent |
| `/delete` | `<id>` | Delete torrent (keep files on disk) |
| `/notify` | `on` or `off` | Enable/disable download-completion notifications per chat |

## Configuration (env vars)

| Variable | Required | Description |
|----------|----------|-------------|
| `TELEGRAM_BOT_TOKEN` | Yes | Bot token from @BotFather |
| `QBITTORRENT_URL` | Yes | qBittorrent Web UI URL, e.g. `http://192.168.1.100:8080` |
| `QBITTORRENT_USER` | No | Web UI username (default: `admin`) |
| `QBITTORRENT_PASS` | No | Web UI password (default: `adminadmin`) |
| `ALLOWED_USERS` | No | Comma-separated Telegram user IDs (empty = allow all) |

## Data Flow

### Adding a torrent

1. User sends magnet link or `.torrent` file to bot
2. Bot authenticates with qBittorrent (login → get cookie)
3. Bot calls `/api/v2/torrents/add` with the magnet URL or file bytes
4. Bot replies with confirmation: "Added: <name>"

### Completion notification

1. Background goroutine polls `/api/v2/torrents/info` every 30s
2. Tracks previously completed torrent hashes
3. When a new "uploading" or "pausedUP" state appears, sends "Download complete: <name>" to all chats with notify enabled

### Querying tasks

1. `/list` → fetches all torrents from qBittorrent
2. Returns formatted list with ID, name, progress bar, size, state

## Project Structure

```
qbittorrent-bot/
├── main.go              # Entrypoint, config, bot setup
├── bot/
│   └── bot.go           # Telegram message handler, command routing
├── qbit/
│   └── qbit.go           # qBittorrent API client (login, add, list, pause, resume, delete)
├── notifier/
│   └── notifier.go       # Completion notification loop
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

## Error Handling

- qBittorrent unreachable → reply "qBittorrent is not reachable"
- Invalid torrent ID → reply "Invalid ID: <id>"
- Auth failure → reply "Failed to authenticate with qBittorrent"
- All errors logged to stdout for docker logs

## Testing

- Unit tests for `qbit/` package using httptest mock server
- Unit tests for command parsing in `bot/`
