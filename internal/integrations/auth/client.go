package auth

import (
	"errors"
	"net/url"
	"strings"
)

type Client struct {
	BaseURL string
}

func NewClient(baseURL string) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")

	if baseURL == "" {
		return nil, errors.New("AUTH_BASE_URL não configurada")
	}

	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Host == "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, errors.New("AUTH_BASE_URL inválida")
	}

	return &Client{
		BaseURL: baseURL,
	}, nil
}
