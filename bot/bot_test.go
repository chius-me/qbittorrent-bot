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

func TestFormatInfo(t *testing.T) {
	info := formatInfo("v4.6.0", &qbit.BuildInfo{
		Libtorrent: "2.0.9.0",
		Qt:         "6.4.2",
		Boost:      "1.83.0",
	})

	if !strings.Contains(info, "v4.6.0") {
		t.Fatalf("expected version in output, got: %s", info)
	}
	if !strings.Contains(info, "2.0.9.0") {
		t.Fatalf("expected libtorrent in output, got: %s", info)
	}
	if !strings.Contains(info, "6.4.2") {
		t.Fatalf("expected Qt in output, got: %s", info)
	}
}

func TestFormatStats(t *testing.T) {
	stats := formatStats(&qbit.TransferInfo{
		DlSpeed:         10485760,
		UpSpeed:         2097152,
		DlData:          1073741824,
		UpData:          536870912,
		DHTNodes:        285,
		ConnectionStatus: "connected",
	}, 15, 3)

	if !strings.Contains(stats, "10.0 MB/s") || !strings.Contains(stats, "2.0 MB/s") {
		t.Fatalf("expected speeds in stats, got: %s", stats)
	}
	if !strings.Contains(stats, "1.0 GB") || !strings.Contains(stats, "512.0 MB") {
		t.Fatalf("expected data totals in stats, got: %s", stats)
	}
	if !strings.Contains(stats, "285") || !strings.Contains(stats, "已连接") {
		t.Fatalf("expected dht/conn in stats, got: %s", stats)
	}
	if !strings.Contains(stats, "3 / 15") {
		t.Fatalf("expected active/total in stats, got: %s", stats)
	}
}
