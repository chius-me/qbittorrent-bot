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

	qbitClient, err := qbit.NewClient(qbitURL, qbitUser, qbitPass)
	if err != nil {
		log.Fatalf("create qBittorrent client: %v", err)
	}
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
