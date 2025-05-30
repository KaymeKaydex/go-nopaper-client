package nopaper

import (
	"crypto/tls"
	"fmt"
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
	// InsecureSkipVerify - ignores ssl certificates,
	// it might be usefully if you have no CA certificates.
	InsecureSkipVerify bool
}

// NewClient creates new Nopaper service client.
func NewClient(cfg Config) (*Client, error) {
	if cfg.Token == "" {
		return nil, fmt.Errorf("token cant be nil")
	}
	// Preparing URL for a work.
	cfg.URL = strings.ReplaceAll(cfg.URL, " ", "")
	cfg.URL = strings.TrimSuffix(cfg.URL, "/")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.InsecureSkipVerify},
	}
	client := &http.Client{Transport: tr}

	client.Transport = newAuthHeaderTransport(tr, cfg.Token)

	return &Client{
		client: client,
		url:    cfg.URL + "/partner-api/api/v2/external",
	}, nil
}
