package bot

import (
	"strings"
	"testing"

	"github.com/chius/qbittorrent-bot/qbit"
)

func TestFormatTorrentListEntryUsesTelegramCardStyle(t *testing.T) {
	torrent := qbit.TorrentInfo{
		Name:     "[hyakuhuyu&LoliHouse] Re Zero kara Hajimeru Isekai Seikatsu - 70 [WebRip 1080p HEVC-10bit AAC ASSx2].mkv",
		Size:     1234567890,
		Progress: 1,
		State:    "stalledUP",
	}

	entry := formatTorrentListEntry(0, torrent)

	if !strings.Contains(entry, "✅ <b>#0</b> 做种等待\n") {
		t.Fatalf("expected title line to start with emoji, id, and state; got:\n%s", entry)
	}
	if strings.Contains(entry, "#0 ✅") {
		t.Fatalf("expected emoji before id, got old order:\n%s", entry)
	}
	if !strings.Contains(entry, "hyakuhuyu&amp;LoliHouse") {
		t.Fatalf("expected torrent name to be HTML escaped, got:\n%s", entry)
	}
	if !strings.Contains(entry, "<code>1.1 GB · 100.0%</code>") {
		t.Fatalf("expected size and progress metadata, got:\n%s", entry)
	}
	if !strings.Contains(entry, "━━━━━━━━━━") {
		t.Fatalf("expected full-width progress divider, got:\n%s", entry)
	}
}
