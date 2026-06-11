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
