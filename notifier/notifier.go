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

type notifierState struct {
	Chats     map[int64]bool `json:"chats"`
	Completed map[string]bool `json:"completed"`
}

type Notifier struct {
	client    *qbit.Client
	bot       *tgbotapi.BotAPI
	mu        sync.Mutex
	state     notifierState
	stateFile string
}

func New(client *qbit.Client, bot *tgbotapi.BotAPI, stateFile string) *Notifier {
	return &Notifier{
		client: client,
		bot:    bot,
		state: notifierState{
			Chats:     make(map[int64]bool),
			Completed: make(map[string]bool),
		},
		stateFile: stateFile,
	}
}

func (n *Notifier) SetNotify(chatID int64, enabled bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if enabled {
		n.state.Chats[chatID] = true
	} else {
		delete(n.state.Chats, chatID)
	}
	n.saveState()
}

func (n *Notifier) LoadState() {
	data, err := os.ReadFile(n.stateFile)
	if err != nil {
		return
	}
	var state notifierState
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}
	n.mu.Lock()
	if state.Chats != nil {
		n.state.Chats = state.Chats
	}
	if state.Completed != nil {
		n.state.Completed = state.Completed
	}
	n.mu.Unlock()
}

func (n *Notifier) saveState() {
	data, err := json.Marshal(n.state)
	if err != nil {
		log.Printf("notifier: failed to marshal state: %v", err)
		return
	}
	if err := os.WriteFile(n.stateFile, data, 0644); err != nil {
		log.Printf("notifier: failed to write state file: %v", err)
	}
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

	activeHashes := make(map[string]bool, len(torrents))
	for _, t := range torrents {
		activeHashes[t.Hash] = true
		if t.Progress >= 1.0 && !n.state.Completed[t.Hash] {
			n.state.Completed[t.Hash] = true
			for chatID := range n.state.Chats {
				msg := tgbotapi.NewMessage(chatID, "下载完成: "+t.Name)
				if _, err := n.bot.Send(msg); err != nil {
					log.Printf("notifier: failed to send to chat %d: %v", chatID, err)
				}
			}
		}
	}

	for hash := range n.state.Completed {
		if !activeHashes[hash] {
			delete(n.state.Completed, hash)
		}
	}
}
