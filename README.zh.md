<p align="right">
  <a href="./README.md">English</a> | 简体中文
</p>

<h1 align="center">qbittorrent-bot</h1>

<p align="center">
  一个 Telegram Bot，用于远程控制 qBittorrent — 发送磁力链接或种子文件即可下载，支持任务管理和完成通知。
</p>

<p align="center">
  <img alt="License" src="https://img.shields.io/badge/License-GPL--3.0-blue.svg">
  <img alt="Language" src="https://img.shields.io/badge/Language-Go-00ADD8.svg">
  <img alt="CI" src="https://github.com/chius-me/qbittorrent-bot/actions/workflows/ci.yml/badge.svg">
  <img alt="Docker" src="https://github.com/chius-me/qbittorrent-bot/actions/workflows/docker.yml/badge.svg">
</p>

---

## ✨ 功能

- **📥 添加下载** — 直接发送磁力链接或 .torrent 文件即可添加任务
- **📋 查看列表** — 显示所有任务的进度条、大小和状态
- **⏯️ 任务控制** — 按序号暂停、恢复或删除任务
- **🔔 完成通知** — 下载完成后自动发送通知
- **🔒 访问控制** — 可限制仅特定 Telegram 用户可用
- **🐳 独立部署** — 独立 Docker 容器，连接任意可达的 qBittorrent

## 快速开始

### 准备工作

- 通过 [@BotFather](https://t.me/BotFather) 创建一个 Telegram Bot，获取 Token
- 一台已启动 Web UI 的 [qBittorrent](https://www.qbittorrent.org/)

### 1. 下载配置

```bash
git clone https://github.com/chius-me/qbittorrent-bot.git
cd qbittorrent-bot
cp .env.example .env
# 编辑 .env 填入你的配置
```

### 2. 启动

```bash
docker compose up -d
```

### 3. 测试

在 Telegram 中给你的 Bot 发送 `/list`。

## 命令列表

| 命令 | 说明 |
|------|------|
| `/add <磁力链接或URL>` | 添加下载任务 |
| *(发送 .torrent 文件)* | 添加种子文件 |
| `/list` | 查看所有下载任务 |
| `/pause <序号>` | 暂停任务（序号来自 /list） |
| `/resume <序号>` | 恢复暂停的任务 |
| `/delete <序号>` | 删除任务（保留文件） |
| `/notify on|off` | 开关下载完成通知 |

## 环境变量

复制 `.env.example` 为 `.env` 并填写：

| 变量 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `TELEGRAM_BOT_TOKEN` | ✅ | — | 从 @BotFather 获取的 Bot Token |
| `QBITTORRENT_URL` | ✅ | — | qBittorrent Web UI 地址（如 `http://192.168.1.100:8080`） |
| `QBITTORRENT_USER` | — | `admin` | Web UI 用户名 |
| `QBITTORRENT_PASS` | — | `adminadmin` | Web UI 密码 |
| `ALLOWED_USERS` | — | *(全部放行)* | 允许使用的 Telegram 用户 ID，多个用逗号分隔 |
| `HTTP_PROXY` | — | — | 用于访问 Telegram API 的 HTTP 代理 |
| `TZ` | — | `Asia/Shanghai` | 容器时区 |

> **注意：** Bot 与 qBittorrent 必须在同一网络可达。如果 qBittorrent 在宿主机上，可使用 `172.17.0.1` 或 `host.docker.internal` 访问。

## Docker 镜像

预构建镜像托管在 GitHub Container Registry：

```text
ghcr.io/chius-me/qbittorrent-bot
```

```bash
# 最新版
docker pull ghcr.io/chius-me/qbittorrent-bot:latest

# 指定版本
docker pull ghcr.io/chius-me/qbittorrent-bot:v1.0.0
```

`latest` 标签在每次推送 `main` 分支时更新。打 `v*` 标签会同时发布版本对应的镜像。

## 直接运行

```bash
go build -o qbittorrent-bot .
export TELEGRAM_BOT_TOKEN=...
export QBITTORRENT_URL=...
./qbittorrent-bot
```

## 测试

```bash
go test ./... -v
```

## 项目结构

- `main.go` — 入口，环境变量配置，组件编排
- `bot/` — Telegram 消息处理和命令路由
- `qbit/` — qBittorrent Web API 客户端
- `notifier/` — 下载完成轮询和通知
- `Dockerfile` — 多阶段 Docker 构建
- `.github/workflows/` — CI 和 Docker 镜像发布

## License

[GPL-3.0](LICENSE)
