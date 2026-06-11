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
