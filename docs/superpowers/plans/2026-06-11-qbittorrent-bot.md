# qBittorrent Telegram Bot — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go Telegram bot that controls qBittorrent — add torrents, manage tasks, and notify on completion.

**Architecture:** Three packages — `qbit` (qBittorrent API client using stdlib HTTP), `bot` (Telegram command handler using go-telegram-bot-api v5.5.1), `notifier` (periodic completion poller). Single `main.go` wires them together. Deployed as standalone Docker container.

**Tech Stack:** Go 1.22+, `github.com/go-telegram-bot-api/telegram-bot-api/v5` v5.5.1, stdlib `net/http`

---

### Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `main.go`

- [ ] **Step 1: Initialize Go module**

Run: `go mod init github.com/chius/qbittorrent-bot`
Expected: creates `go.mod`

- [ ] **Step 2: Add telegram-bot-api dependency**

Run: `go get github.com/go-telegram-bot-api/telegram-bot-api/v5@v5.5.1`
Expected: updates `go.mod` and creates `go.sum`

- [ ] **Step 3: Create minimal main.go**

```go
package main

import (
	"log"
	"os"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}
	log.Println("Bot starting...")
}
```

- [ ] **Step 4: Verify it compiles**

Run: `go build -o /dev/null ./...`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum main.go
git commit -m "chore: scaffold Go project with dependencies"
```

---

### Task 2: qBittorrent API Client — Login and Types

**Files:**
- Create: `qbit/qbit.go`
- Create: `qbit/qbit_test.go`

- [ ] **Step 1: Write failing test for Client.Login**

```go
package qbit

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/auth/login" && r.Method == "POST" {
			r.ParseForm()
			if r.FormValue("username") == "admin" && r.FormValue("password") == "adminadmin" {
				http.SetCookie(w, &http.Cookie{Name: "SID", Value: "test-sid"})
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(server.URL, "admin", "adminadmin")
	err := c.Login()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./qbit/ -v -run TestLogin`
Expected: FAIL (undefined: NewClient)

- [ ] **Step 3: Implement Client struct and NewClient**

```go
package qbit

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

func NewClient(baseURL, username, password string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		baseURL:    baseURL,
		username:   username,
		password:   password,
		httpClient: &http.Client{Jar: jar},
	}
}
```

- [ ] **Step 4: Implement Login method**

```go
func (c *Client) Login() error {
	resp, err := c.httpClient.PostForm(c.baseURL+"/api/v2/auth/login", url.Values{
		"username": {c.username},
		"password": {c.password},
	})
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d", resp.StatusCode)
	}
	return nil
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./qbit/ -v -run TestLogin`
Expected: PASS

- [ ] **Step 6: Write test for login failure**

```go
func TestLoginFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	c := NewClient(server.URL, "bad", "wrong")
	err := c.Login()
	if err == nil {
		t.Fatal("expected error for bad credentials")
	}
}
```

- [ ] **Step 7: Run test**

Run: `go test ./qbit/ -v -run TestLoginFailure`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add qbit/
git commit -m "feat: qBittorrent client with login"
```

---

### Task 3: qBittorrent API Client — Add Torrents

**Files:**
- Modify: `qbit/qbit.go`
- Modify: `qbit/qbit_test.go`

- [ ] **Step 1: Write test for AddMagnet**

```go
func TestAddMagnet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/torrents/add" && r.Method == "POST" {
			r.ParseForm()
			if r.FormValue("urls") == "magnet:?xt=urn:btih:abc" {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(server.URL, "", "")
	err := c.AddMagnet("magnet:?xt=urn:btih:abc")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./qbit/ -v -run TestAddMagnet`
Expected: FAIL (undefined: AddMagnet)

- [ ] **Step 3: Implement AddMagnet**

```go
func (c *Client) AddMagnet(magnet string) error {
	resp, err := c.httpClient.PostForm(c.baseURL+"/api/v2/torrents/add", url.Values{
		"urls": {magnet},
	})
	if err != nil {
		return fmt.Errorf("add magnet request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("add magnet failed with status %d", resp.StatusCode)
	}
	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./qbit/ -v -run TestAddMagnet`
Expected: PASS

- [ ] **Step 5: Write test for AddTorrentFile**

```go
func TestAddTorrentFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/torrents/add" && r.Method == "POST" {
			err := r.ParseMultipartForm(10 << 20)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			file, _, err := r.FormFile("torrents")
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer file.Close()
			buf := make([]byte, 1024)
			n, _ := file.Read(buf)
			if string(buf[:n]) == "torrent-content" {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(server.URL, "", "")
	err := c.AddTorrentFile("test.torrent", []byte("torrent-content"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
```

- [ ] **Step 6: Run test to verify it fails**

Run: `go test ./qbit/ -v -run TestAddTorrentFile`
Expected: FAIL

- [ ] **Step 7: Implement AddTorrentFile**

```go
func (c *Client) AddTorrentFile(filename string, data []byte) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("torrents", filename)
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return fmt.Errorf("write torrent data: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", c.baseURL+"/api/v2/torrents/add", body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("add torrent file request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("add torrent file failed with status %d", resp.StatusCode)
	}
	return nil
}
```

Add imports to `qbit.go`:
```go
import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)
```

- [ ] **Step 8: Run tests**

Run: `go test ./qbit/ -v -run TestAdd`
Expected: both PASS

- [ ] **Step 9: Commit**

```bash
git add qbit/
git commit -m "feat: add magnet and torrent file methods"
```

---

### Task 4: qBittorrent API Client — List and Manage Torrents

**Files:**
- Modify: `qbit/qbit.go`
- Modify: `qbit/qbit_test.go`

- [ ] **Step 1: Define TorrentInfo type and write List test**

Add to `qbit.go`:
```go
type TorrentInfo struct {
	Hash      string  `json:"hash"`
	Name      string  `json:"name"`
	Size      int64   `json:"size"`
	Progress  float64 `json:"progress"`
	State     string  `json:"state"`
	Dlspeed   int64   `json:"dlspeed"`
	AddedOn   int64   `json:"added_on"`
	Category  string  `json:"category"`
}
```

Add to `qbit_test.go`:
```go
func TestList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/torrents/info" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[{"hash":"abc","name":"test","size":1024,"progress":0.5,"state":"downloading","dlspeed":100,"added_on":1000,"category":""}]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(server.URL, "", "")
	torrents, err := c.List()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(torrents) != 1 {
		t.Fatalf("expected 1 torrent, got %d", len(torrents))
	}
	if torrents[0].Name != "test" {
		t.Fatalf("expected name 'test', got '%s'", torrents[0].Name)
	}
	if torrents[0].Progress != 0.5 {
		t.Fatalf("expected progress 0.5, got %f", torrents[0].Progress)
	}
}
```

- [ ] **Step 2: Verify test fails**

Run: `go test ./qbit/ -v -run TestList`
Expected: FAIL

- [ ] **Step 3: Implement List**

Add to `qbit.go`:
```go
func (c *Client) List() ([]TorrentInfo, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/v2/torrents/info")
	if err != nil {
		return nil, fmt.Errorf("list request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list failed with status %d", resp.StatusCode)
	}

	var torrents []TorrentInfo
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return torrents, nil
}
```

Add `"encoding/json"` and `"io"` to imports (remove `io` if already imported and keep as `"encoding/json"`).

- [ ] **Step 4: Run test**

Run: `go test ./qbit/ -v -run TestList`
Expected: PASS

- [ ] **Step 5: Write test for Pause, Resume, Delete**

```go
func TestPauseResumeDelete(t *testing.T) {
	lastPath := ""
	lastHashes := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		lastPath = r.URL.Path
		lastHashes = r.FormValue("hashes")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "", "")

	c.Pause("abc")
	if lastPath != "/api/v2/torrents/pause" || lastHashes != "abc" {
		t.Fatalf("pause: path=%s hashes=%s", lastPath, lastHashes)
	}

	c.Resume("def")
	if lastPath != "/api/v2/torrents/resume" || lastHashes != "def" {
		t.Fatalf("resume: path=%s hashes=%s", lastPath, lastHashes)
	}

	c.Delete("ghi", false)
	if lastPath != "/api/v2/torrents/delete" || lastHashes != "ghi" {
		t.Fatalf("delete: path=%s hashes=%s", lastPath, lastHashes)
	}
}
```

- [ ] **Step 6: Verify test fails**

Run: `go test ./qbit/ -v -run TestPauseResumeDelete`
Expected: FAIL

- [ ] **Step 7: Implement Pause, Resume, Delete**

```go
func (c *Client) Pause(hash string) error {
	return c.torrentAction("pause", hash)
}

func (c *Client) Resume(hash string) error {
	return c.torrentAction("resume", hash)
}

func (c *Client) Delete(hash string, deleteFiles bool) error {
	deleteFilesStr := "false"
	if deleteFiles {
		deleteFilesStr = "true"
	}
	resp, err := c.httpClient.PostForm(c.baseURL+"/api/v2/torrents/delete", url.Values{
		"hashes":      {hash},
		"deleteFiles": {deleteFilesStr},
	})
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed with status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) torrentAction(action, hash string) error {
	resp, err := c.httpClient.PostForm(c.baseURL+"/api/v2/torrents/"+action, url.Values{
		"hashes": {hash},
	})
	if err != nil {
		return fmt.Errorf("%s request failed: %w", action, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s failed with status %d", action, resp.StatusCode)
	}
	return nil
}
```

- [ ] **Step 8: Run all tests**

Run: `go test ./qbit/ -v`
Expected: all PASS

- [ ] **Step 9: Commit**

```bash
git add qbit/
git commit -m "feat: list, pause, resume, delete torrent methods"
```

---

### Task 5: qBittorrent API Client — File Import Cleanup

**Files:**
- Modify: `qbit/qbit.go`

- [ ] **Step 1: Ensure clean imports in qbit.go**

The final import block should be:
```go
import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)
```

- [ ] **Step 2: Verify compiles**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 3: Run full test suite**

Run: `go test ./qbit/ -v`
Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add qbit/
git commit -m "chore: clean up imports in qbit package"
```

---

### Task 6: Notifier — Completion Polling

**Files:**
- Create: `notifier/notifier.go`

- [ ] **Step 1: Create notifier package**

```go
package notifier

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/chius/qbittorrent-bot/qbit"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Notifier struct {
	client    *qbit.Client
	bot       *tgbotapi.BotAPI
	mu        sync.Mutex
	chats     map[int64]bool
	completed map[string]bool
	stateFile string
}

func New(client *qbit.Client, bot *tgbotapi.BotAPI, stateFile string) *Notifier {
	return &Notifier{
		client:    client,
		bot:       bot,
		chats:     make(map[int64]bool),
		completed: make(map[string]bool),
		stateFile: stateFile,
	}
}
```

- [ ] **Step 2: Add SetNotify and load/save state**

```go
func (n *Notifier) SetNotify(chatID int64, enabled bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if enabled {
		n.chats[chatID] = true
	} else {
		delete(n.chats, chatID)
	}
	n.saveState()
}

func (n *Notifier) LoadState() {
	data, err := os.ReadFile(n.stateFile)
	if err != nil {
		return
	}
	var state struct {
		Chats map[int64]bool `json:"chats"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}
	n.mu.Lock()
	n.chats = state.Chats
	n.mu.Unlock()
}

func (n *Notifier) saveState() {
	state := struct {
		Chats map[int64]bool `json:"chats"`
	}{Chats: n.chats}
	data, _ := json.Marshal(state)
	os.WriteFile(n.stateFile, data, 0644)
}
```

- [ ] **Step 3: Add Start polling loop**

```go
func (n *Notifier) Start() {
	n.LoadState()

	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			n.checkCompletions()
		}
	}()
}

func (n *Notifier) checkCompletions() {
	torrents, err := n.client.List()
	if err != nil {
		log.Printf("notifier: failed to list torrents: %v", err)
		return
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	for _, t := range torrents {
		if t.Progress >= 1.0 && !n.completed[t.Hash] {
			n.completed[t.Hash] = true
			for chatID := range n.chats {
				msg := tgbotapi.NewMessage(chatID, "下载完成: "+t.Name)
				if _, err := n.bot.Send(msg); err != nil {
					log.Printf("notifier: failed to send to chat %d: %v", chatID, err)
				}
			}
		}
	}
}
```

- [ ] **Step 4: Verify compiles**

Run: `go build ./...`
Expected: might need `go mod tidy` first

- [ ] **Step 5: Commit**

```bash
go mod tidy
git add notifier/ go.mod go.sum
git commit -m "feat: completion notifier with per-chat state"
```

---

### Task 7: Bot — Command Handler

**Files:**
- Create: `bot/bot.go`

- [ ] **Step 1: Create bot package with Bot struct and command routing**

```go
package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/chius/qbittorrent-bot/qbit"
	"github.com/chius/qbittorrent-bot/notifier"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api      *tgbotapi.BotAPI
	qbit     *qbit.Client
	notifier *notifier.Notifier
	allowed  map[int64]bool
}

func New(token string, qbitClient *qbit.Client, notif *notifier.Notifier, allowedUsers string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	allowed := make(map[int64]bool)
	if allowedUsers != "" {
		for _, s := range strings.Split(allowedUsers, ",") {
			s = strings.TrimSpace(s)
			id, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid allowed user id: %s", s)
			}
			allowed[id] = true
		}
	}

	return &Bot{
		api:      api,
		qbit:     qbitClient,
		notifier: notif,
		allowed:  allowed,
	}, nil
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		b.handleMessage(update.Message)
	}
}
```

- [ ] **Step 2: Add access check and message router**

```go
func (b *Bot) isAllowed(chatID int64) bool {
	if len(b.allowed) == 0 {
		return true
	}
	return b.allowed[chatID]
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	if !b.isAllowed(msg.Chat.ID) {
		b.reply(msg.Chat.ID, msg.MessageID, "你没有权限使用此 bot")
		return
	}

	if msg.IsCommand() {
		b.handleCommand(msg)
	} else if msg.Document != nil && strings.HasSuffix(msg.Document.FileName, ".torrent") {
		b.handleTorrentFile(msg)
	} else if msg.Text != "" {
		b.handleText(msg)
	}
}

func (b *Bot) reply(chatID int64, replyTo int, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = replyTo
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("failed to send message: %v", err)
	}
}

func (b *Bot) replyHTML(chatID int64, replyTo int, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = replyTo
	msg.ParseMode = "HTML"
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("failed to send message: %v", err)
	}
}
```

- [ ] **Step 3: Add command handlers**

```go
func (b *Bot) handleCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "add":
		b.cmdAdd(msg)
	case "list":
		b.cmdList(msg)
	case "pause":
		b.cmdPause(msg)
	case "resume":
		b.cmdResume(msg)
	case "delete":
		b.cmdDelete(msg)
	case "notify":
		b.cmdNotify(msg)
	default:
		b.reply(msg.Chat.ID, msg.MessageID, "未知命令。可用命令: /add, /list, /pause, /resume, /delete, /notify")
	}
}
```

- [ ] **Step 4: Implement /add command**

```go
func (b *Bot) cmdAdd(msg *tgbotapi.Message) {
	link := strings.TrimSpace(msg.CommandArguments())
	if link == "" {
		b.reply(msg.Chat.ID, msg.MessageID, "用法: /add <磁力链接或URL>")
		return
	}

	if err := b.qbit.AddMagnet(link); err != nil {
		log.Printf("add magnet failed: %v", err)
		b.reply(msg.Chat.ID, msg.MessageID, "添加下载失败，请检查链接格式")
		return
	}
	b.reply(msg.Chat.ID, msg.MessageID, "已添加下载任务")
}
```

- [ ] **Step 5: Implement /list command**

```go
func (b *Bot) cmdList(msg *tgbotapi.Message) {
	torrents, err := b.qbit.List()
	if err != nil {
		log.Printf("list failed: %v", err)
		b.reply(msg.Chat.ID, msg.MessageID, "获取任务列表失败")
		return
	}

	if len(torrents) == 0 {
		b.reply(msg.Chat.ID, msg.MessageID, "当前没有下载任务")
		return
	}

	var sb strings.Builder
	for i, t := range torrents {
		progressBar := buildProgressBar(t.Progress)
		sizeStr := formatSize(t.Size)
		stateStr := translateState(t.State)
		sb.WriteString(fmt.Sprintf("<b>[%d]</b> %s\n", i, t.Name))
		sb.WriteString(fmt.Sprintf("    %s %.1f%%  %s  %s\n",
			progressBar, t.Progress*100, sizeStr, stateStr))
	}
	b.replyHTML(msg.Chat.ID, msg.MessageID, sb.String())
}

func buildProgressBar(progress float64) string {
	filled := int(progress * 10)
	if filled > 10 {
		filled = 10
	}
	empty := 10 - filled
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func translateState(state string) string {
	switch state {
	case "downloading":
		return "下载中"
	case "uploading", "forcedUP":
		return "做种中"
	case "pausedUP":
		return "已暂停(完成)"
	case "pausedDL":
		return "已暂停"
	case "queuedDL", "queuedUP":
		return "排队中"
	case "stalledDL":
		return "等待中"
	case "stalledUP":
		return "做种等待"
	case "checkingDL", "checkingUP":
		return "校验中"
	case "allocating":
		return "分配中"
	case "metaDL":
		return "获取元数据"
	case "missingFiles":
		return "文件缺失"
	case "error":
		return "错误"
	case "moving":
		return "移动中"
	default:
		return state
	}
}
```

- [ ] **Step 6: Implement /pause, /resume, /delete commands**

```go
func (b *Bot) cmdPause(msg *tgbotapi.Message) {
	args := strings.TrimSpace(msg.CommandArguments())
	if args == "" {
		b.reply(msg.Chat.ID, msg.MessageID, "用法: /pause <任务序号>\n先用 /list 查看序号")
		return
	}

	torrents, err := b.qbit.List()
	if err != nil {
		b.reply(msg.Chat.ID, msg.MessageID, "获取任务列表失败")
		return
	}

	idx, err := strconv.Atoi(args)
	if err != nil || idx < 0 || idx >= len(torrents) {
		b.reply(msg.Chat.ID, msg.MessageID, "无效的任务序号")
		return
	}

	if err := b.qbit.Pause(torrents[idx].Hash); err != nil {
		log.Printf("pause failed: %v", err)
		b.reply(msg.Chat.ID, msg.MessageID, "暂停失败")
		return
	}
	b.reply(msg.Chat.ID, msg.MessageID, "已暂停: "+torrents[idx].Name)
}

func (b *Bot) cmdResume(msg *tgbotapi.Message) {
	args := strings.TrimSpace(msg.CommandArguments())
	if args == "" {
		b.reply(msg.Chat.ID, msg.MessageID, "用法: /resume <任务序号>\n先用 /list 查看序号")
		return
	}

	torrents, err := b.qbit.List()
	if err != nil {
		b.reply(msg.Chat.ID, msg.MessageID, "获取任务列表失败")
		return
	}

	idx, err := strconv.Atoi(args)
	if err != nil || idx < 0 || idx >= len(torrents) {
		b.reply(msg.Chat.ID, msg.MessageID, "无效的任务序号")
		return
	}

	if err := b.qbit.Resume(torrents[idx].Hash); err != nil {
		log.Printf("resume failed: %v", err)
		b.reply(msg.Chat.ID, msg.MessageID, "恢复失败")
		return
	}
	b.reply(msg.Chat.ID, msg.MessageID, "已恢复: "+torrents[idx].Name)
}

func (b *Bot) cmdDelete(msg *tgbotapi.Message) {
	args := strings.TrimSpace(msg.CommandArguments())
	if args == "" {
		b.reply(msg.Chat.ID, msg.MessageID, "用法: /delete <任务序号>\n先用 /list 查看序号")
		return
	}

	torrents, err := b.qbit.List()
	if err != nil {
		b.reply(msg.Chat.ID, msg.MessageID, "获取任务列表失败")
		return
	}

	idx, err := strconv.Atoi(args)
	if err != nil || idx < 0 || idx >= len(torrents) {
		b.reply(msg.Chat.ID, msg.MessageID, "无效的任务序号")
		return
	}

	if err := b.qbit.Delete(torrents[idx].Hash, false); err != nil {
		log.Printf("delete failed: %v", err)
		b.reply(msg.Chat.ID, msg.MessageID, "删除失败")
		return
	}
	b.reply(msg.Chat.ID, msg.MessageID, "已删除: "+torrents[idx].Name)
}
```

- [ ] **Step 7: Implement /notify command and file/torrent handlers**

```go
func (b *Bot) cmdNotify(msg *tgbotapi.Message) {
	args := strings.TrimSpace(msg.CommandArguments())
	switch args {
	case "on":
		b.notifier.SetNotify(msg.Chat.ID, true)
		b.reply(msg.Chat.ID, msg.MessageID, "下载完成通知已开启")
	case "off":
		b.notifier.SetNotify(msg.Chat.ID, false)
		b.reply(msg.Chat.ID, msg.MessageID, "下载完成通知已关闭")
	default:
		b.reply(msg.Chat.ID, msg.MessageID, "用法: /notify on|off")
	}
}

func (b *Bot) handleText(msg *tgbotapi.Message) {
	text := strings.TrimSpace(msg.Text)
	if strings.HasPrefix(text, "magnet:") || strings.HasPrefix(text, "http") {
		b.cmdAdd(msg)
	} else {
		b.reply(msg.Chat.ID, msg.MessageID, "请发送磁力链接、种子文件，或使用命令: /add, /list, /pause, /resume, /delete, /notify")
	}
}

func (b *Bot) handleTorrentFile(msg *tgbotapi.Message) {
	file := msg.Document
	if !strings.HasSuffix(file.FileName, ".torrent") {
		b.reply(msg.Chat.ID, msg.MessageID, "请发送 .torrent 文件")
		return
	}

	fileCfg := tgbotapi.FileConfig{FileID: file.FileID}
	f, err := b.api.GetFile(fileCfg)
	if err != nil {
		log.Printf("get file failed: %v", err)
		b.reply(msg.Chat.ID, msg.MessageID, "获取文件失败")
		return
	}

	fileURL := f.Link(b.api.Token)
	resp, err := http.Get(fileURL)
	if err != nil {
		log.Printf("download file failed: %v", err)
		b.reply(msg.Chat.ID, msg.MessageID, "下载种子文件失败")
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("read file failed: %v", err)
		b.reply(msg.Chat.ID, msg.MessageID, "读取种子文件失败")
		return
	}

	if err := b.qbit.AddTorrentFile(file.FileName, data); err != nil {
		log.Printf("add torrent file failed: %v", err)
		b.reply(msg.Chat.ID, msg.MessageID, "添加种子失败")
		return
	}
	b.reply(msg.Chat.ID, msg.MessageID, "已添加种子: "+file.FileName)
}
```

Add imports to `bot.go`:
```go
import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/chius/qbittorrent-bot/qbit"
	"github.com/chius/qbittorrent-bot/notifier"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)
```

- [ ] **Step 8: Verify compiles**

Run: `go build ./...`
Expected: no errors (may need `go mod tidy`)

- [ ] **Step 9: Commit**

```bash
go mod tidy
git add bot/ go.mod go.sum
git commit -m "feat: telegram bot command handler"
```

---

### Task 8: Main Entry Point and Wiring

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Write full main.go**

```go
package main

import (
	"log"
	"os"

	"github.com/chius/qbittorrent-bot/bot"
	"github.com/chius/qbittorrent-bot/notifier"
	"github.com/chius/qbittorrent-bot/qbit"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	qbitURL := os.Getenv("QBITTORRENT_URL")
	if qbitURL == "" {
		log.Fatal("QBITTORRENT_URL is required")
	}

	qbitUser := os.Getenv("QBITTORRENT_USER")
	if qbitUser == "" {
		qbitUser = "admin"
	}

	qbitPass := os.Getenv("QBITTORRENT_PASS")
	if qbitPass == "" {
		qbitPass = "adminadmin"
	}

	allowedUsers := os.Getenv("ALLOWED_USERS")
	stateFile := os.Getenv("STATE_FILE")
	if stateFile == "" {
		stateFile = "/data/notify_state.json"
	}

	qbitClient := qbit.NewClient(qbitURL, qbitUser, qbitPass)
	if err := qbitClient.Login(); err != nil {
		log.Fatalf("qBittorrent login failed: %v", err)
	}
	log.Println("qBittorrent login successful")

	b, err := bot.New(token, qbitClient, nil, allowedUsers)
	if err != nil {
		log.Fatalf("create bot failed: %v", err)
	}

	notif := notifier.New(qbitClient, b.API(), stateFile)
	notif.Start()
	b.SetNotifier(notif)

	log.Println("Bot started, polling for updates...")
	b.Start()
}
```

- [ ] **Step 2: Expose API() and SetNotifier in bot**

Add to `bot/bot.go`:
```go
func (b *Bot) API() *tgbotapi.BotAPI {
	return b.api
}

func (b *Bot) SetNotifier(n *notifier.Notifier) {
	b.notifier = n
}
```

- [ ] **Step 3: Verify compiles and builds**

Run: `go build -o qbittorrent-bot .`
Expected: no errors, binary created

- [ ] **Step 4: Commit**

```bash
git add main.go bot/bot.go
git commit -m "feat: wire up main entry point"
```

---

### Task 9: Docker Deployment

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`

- [ ] **Step 1: Write Dockerfile**

```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o qbittorrent-bot .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/qbittorrent-bot .

ENV STATE_FILE=/data/notify_state.json
VOLUME ["/data"]

ENTRYPOINT ["./qbittorrent-bot"]
```

- [ ] **Step 2: Write docker-compose.yml**

```yaml
services:
  qbittorrent-bot:
    build: .
    container_name: qbittorrent-bot
    restart: unless-stopped
    environment:
      - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
      - QBITTORRENT_URL=${QBITTORRENT_URL:-http://192.168.1.100:8080}
      - QBITTORRENT_USER=${QBITTORRENT_USER:-admin}
      - QBITTORRENT_PASS=${QBITTORRENT_PASS:-adminadmin}
      - ALLOWED_USERS=${ALLOWED_USERS:-}
      - TZ=${TZ:-Asia/Shanghai}
    volumes:
      - qbittorrent-bot-data:/data

volumes:
  qbittorrent-bot-data:
```

- [ ] **Step 3: Commit**

```bash
git add Dockerfile docker-compose.yml
git commit -m "feat: docker deployment"
```

---

### Task 10: Final Verification

- [ ] **Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: all tests PASS

- [ ] **Step 2: Verify build**

Run: `go build -o qbittorrent-bot .`
Expected: binary created

- [ ] **Step 3: Run go vet**

Run: `go vet ./...`
Expected: no errors

- [ ] **Step 4: Update README**

```markdown
# qbittorrent-bot

Telegram bot for controlling qBittorrent. Send magnet links or torrent files, manage downloads, get notified on completion.

## Setup

1. Create a bot via [@BotFather](https://t.me/BotFather), get the token
2. Copy `.env.example` to `.env` and fill in values
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
```

- [ ] **Step 5: Commit**

```bash
git add README.md
git commit -m "docs: update README with usage instructions"
```
