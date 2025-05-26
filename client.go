package go_nopaper_client

import (
	"net/http"
	"strings"
)

// Client - nopaper service external client.
type Client struct {
	client *http.Client

	url string
}

// Config for nopaper client.
type Config struct {
	// URL is an url of Nopaper service.
	// There is https://np-demo.abanking.ru/ for demo stand as example.
	// It must exclude /partner-api/api/v2/external suffix.
	URL string `yaml:"url"`
	// Token is a Nopaper auth token, that contains in request header X-API-KEY.
	Token string `yaml:"token"`
}

// NewClient creates new Nopaper service client.
func NewClient(cfg Config) (*Client, error) {
	// Preparing URL for a work.
	cfg.URL = strings.ReplaceAll(cfg.URL, " ", "")
	cfg.URL = strings.TrimSuffix(cfg.URL, "/")

	// Provide auth header with token for all client requests.
	client := http.DefaultClient
	client.Transport = newAuthHeaderTransport(http.DefaultTransport, cfg.Token)

	return &Client{
		client: client,
		url:    cfg.URL + "/partner-api/api/v2/external",
	}, nil
}
