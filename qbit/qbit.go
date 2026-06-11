package qbit

import (
	"bytes"
	"fmt"
	"mime/multipart"
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
