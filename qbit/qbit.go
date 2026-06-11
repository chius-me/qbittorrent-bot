package qbit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type TorrentInfo struct {
	Hash     string  `json:"hash"`
	Name     string  `json:"name"`
	Size     int64   `json:"size"`
	Progress float64 `json:"progress"`
	State    string  `json:"state"`
	Dlspeed  int64   `json:"dlspeed"`
	AddedOn  int64   `json:"added_on"`
	Category string  `json:"category"`
}

type BuildInfo struct {
	Libtorrent  string `json:"libtorrent"`
	Qt          string `json:"qt"`
	Boost       string `json:"boost"`
	OpenSSL     string `json:"openssl"`
	Bitness     int    `json:"bitness"`
}

type TransferInfo struct {
	DlSpeed         int64  `json:"dl_info_speed"`
	UpSpeed         int64  `json:"up_info_speed"`
	DlData          int64  `json:"dl_info_data"`
	UpData          int64  `json:"up_info_data"`
	DlRateLimit     int64  `json:"dl_rate_limit"`
	UpRateLimit     int64  `json:"up_rate_limit"`
	DHTNodes        int    `json:"dht_nodes"`
	ConnectionStatus string `json:"connection_status"`
}

type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

func NewClient(baseURL, username, password string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}
	return &Client{
		baseURL:    baseURL,
		username:   username,
		password:   password,
		httpClient: &http.Client{Jar: jar},
	}, nil
}

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

func (c *Client) GetAppVersion() (string, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/v2/app/version")
	if err != nil {
		return "", fmt.Errorf("version request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("version failed with status %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read version: %w", err)
	}
	return strings.TrimSpace(string(b)), nil
}

func (c *Client) GetBuildInfo() (*BuildInfo, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/v2/app/buildInfo")
	if err != nil {
		return nil, fmt.Errorf("build info request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("build info failed with status %d", resp.StatusCode)
	}
	var info BuildInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode build info: %w", err)
	}
	return &info, nil
}

func (c *Client) GetTransferInfo() (*TransferInfo, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/v2/transfer/info")
	if err != nil {
		return nil, fmt.Errorf("transfer info request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("transfer info failed with status %d", resp.StatusCode)
	}
	var info TransferInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode transfer info: %w", err)
	}
	return &info, nil
}

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
