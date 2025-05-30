package nopaper

import (
	"context"
	"net/http"
	"net/url"
)

// SetCallbackURI method sets callback URI for nopaper.
func (c *Client) SetCallbackURI(ctx context.Context, callbackURL string) error {
	v := url.Values{}

	v.Add("uri", callbackURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url+"/hub/callback-uri", nil)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	req.URL.RawQuery = v.Encode()
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	return emptyResponseResult(resp)
}
