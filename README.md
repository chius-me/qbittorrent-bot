# qbittorrent-bot

Telegram bot for controlling qBittorrent. Send magnet links or torrent files, manage downloads, get notified on completion.

## Setup

1. Create a bot via [@BotFather](https://t.me/BotFather), get the token
2. Set environment variables (see below)
3. `docker compose up -d`

## Commands

| Command | Description |
|---------|-------------|
| `/add <link>` | Add magnet/URL |
| (send .torrent file) | Add torrent file |
| `/list` | List all torrents |
| `/pause <id>` | Pause torrent |
| `/resume <id>` | Resume torrent |
| `/delete <id>` | Delete torrent |
| `/notify on/off` | Toggle completion notifications |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TELEGRAM_BOT_TOKEN` | (required) | Bot token |
| `QBITTORRENT_URL` | (required) | qBittorrent Web UI URL |
| `QBITTORRENT_USER` | `admin` | Web UI username |
| `QBITTORRENT_PASS` | `adminadmin` | Web UI password |
| `ALLOWED_USERS` | (all) | Comma-separated Telegram user IDs |
| `TZ` | `Asia/Shanghai` | Timezone |
