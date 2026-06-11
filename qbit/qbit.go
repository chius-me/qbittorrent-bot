package qbit

import (
	"fmt"
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

func NewClient(baseURL, username, password string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		baseURL:    baseURL,
		username:   username,
		password:   password,
		httpClient: &http.Client{Jar: jar},
	}
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
