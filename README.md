<p align="right">
  English | <a href="./README.zh.md">简体中文</a>
</p>

<h1 align="center">qbittorrent-bot</h1>

<p align="center">
  A Telegram bot for controlling qBittorrent — send magnet links or torrent files, manage downloads, get notified on completion.
</p>

<p align="center">
  <img alt="License" src="https://img.shields.io/badge/License-GPL--3.0-blue.svg">
  <img alt="Language" src="https://img.shields.io/badge/Language-Go-00ADD8.svg">
  <img alt="CI" src="https://github.com/chius-me/qbittorrent-bot/actions/workflows/ci.yml/badge.svg">
  <img alt="Docker" src="https://github.com/chius-me/qbittorrent-bot/actions/workflows/docker.yml/badge.svg">
</p>

---

## ✨ Features

- **📥 Add torrents** — Send a magnet link or `.torrent` file directly in chat
- **📋 List tasks** — View all torrents with progress bars, size, and status
- **⏯️ Control downloads** — Pause, resume, or delete torrents by index
- **🔔 Completion notifications** — Get notified when a download finishes
- **🔒 Access control** — Restrict to specific Telegram users via whitelist
- **🐳 Deploy anywhere** — Standalone Docker container, connects to any reachable qBittorrent instance

## Quick Start

### Prerequisites

- A Telegram bot token from [@BotFather](https://t.me/BotFather)
- A running [qBittorrent](https://www.qbittorrent.org/) instance with Web UI enabled

### 1. Clone and configure

```bash
git clone https://github.com/chius-me/qbittorrent-bot.git
cd qbittorrent-bot
cp .env.example .env
# Edit .env with your credentials
```

### 2. Start with Docker

```bash
docker compose up -d
```

### 3. Test your bot

Open Telegram, send `/list` to your bot.

## Commands

| Command | Description |
|---------|-------------|
| `/add <magnet or URL>` | Add a magnet link or HTTP download |
| *(send .torrent file)* | Add a torrent file |
| `/list` | List all torrents with progress |
| `/pause <id>` | Pause a torrent (use index from /list) |
| `/resume <id>` | Resume a paused torrent |
| `/delete <id>` | Delete a torrent (keeps files on disk) |
| `/notify on|off` | Toggle download completion notifications |

## Configuration

The bot is configured via environment variables. Copy `.env.example` to `.env` and fill in:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `TELEGRAM_BOT_TOKEN` | ✅ | — | Bot token from @BotFather |
| `QBITTORRENT_URL` | ✅ | — | qBittorrent Web UI URL (e.g. `http://192.168.1.100:8080`) |
| `QBITTORRENT_USER` | — | `admin` | Web UI username |
| `QBITTORRENT_PASS` | — | `adminadmin` | Web UI password |
| `ALLOWED_USERS` | — | *(all)* | Comma-separated Telegram user IDs |
| `HTTP_PROXY` | — | — | Proxy for Telegram API (e.g. `http://10.10.10.1:7897`) |
| `TZ` | — | `Asia/Shanghai` | Container timezone |

> **Note:** `QBITTORRENT_URL` uses the same network as the bot container. If qBittorrent runs on the Docker host, use `172.17.0.1` or `host.docker.internal` as the host IP.

## Docker Images

Pre-built images are available on GitHub Container Registry:

```text
ghcr.io/chius-me/qbittorrent-bot
```

```bash
# Latest
docker pull ghcr.io/chius-me/qbittorrent-bot:latest

# Specific version
docker pull ghcr.io/chius-me/qbittorrent-bot:v1.0.0
```

The `latest` tag is updated on every push to the `main` branch. Tagged releases (e.g. `v1.0.0`) publish version-specific images.

## Run without Docker

```bash
# Build
go build -o qbittorrent-bot .

# Run (set env vars first)
export TELEGRAM_BOT_TOKEN=...
export QBITTORRENT_URL=...
./qbittorrent-bot
```

## Tests

```bash
go test ./... -v
```

## Project Structure

- `main.go` — Entrypoint, config from env, wiring
- `bot/` — Telegram message handler and command routing
- `qbit/` — qBittorrent Web API client
- `notifier/` — Download completion polling and notifications
- `Dockerfile` — Multi-stage Docker build
- `.github/workflows/` — CI and Docker image publishing

## License

[GPL-3.0](LICENSE)
