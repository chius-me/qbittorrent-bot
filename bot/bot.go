package bot

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/chius/qbittorrent-bot/notifier"
	"github.com/chius/qbittorrent-bot/qbit"
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

func (b *Bot) API() *tgbotapi.BotAPI {
	return b.api
}

func (b *Bot) SetNotifier(n *notifier.Notifier) {
	b.notifier = n
}
