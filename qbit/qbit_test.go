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

	c, err := NewClient(server.URL, "admin", "adminadmin")
	if err != nil {
		t.Fatalf("expected no error from NewClient, got %v", err)
	}
	err = c.Login()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestLoginFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	c, err := NewClient(server.URL, "bad", "wrong")
	if err != nil {
		t.Fatalf("expected no error from NewClient, got %v", err)
	}
	err = c.Login()
	if err == nil {
		t.Fatal("expected error for bad credentials")
	}
}

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

	c, err := NewClient(server.URL, "", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	err = c.AddMagnet("magnet:?xt=urn:btih:abc")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

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

	c, err := NewClient(server.URL, "", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	err = c.AddTorrentFile("test.torrent", []byte("torrent-content"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

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

	c, err := NewClient(server.URL, "", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
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

	c, err := NewClient(server.URL, "", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

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
