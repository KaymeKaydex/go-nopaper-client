package go_nopaper_client

import (
	"fmt"
	"io"
	"net/http"
)

// emptyResponseResult is a helper for empty responses from nopaper.
// It reduces count of duplicate code with simple response code and errors wrapping.
func emptyResponseResult(r *http.Response) error {
	if r.StatusCode == http.StatusOK || r.StatusCode == http.StatusCreated {
		// Good response.
		return nil
	}

	bts, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	return fmt.Errorf("unknown status code from nopaper: %s : %s", r.Status, string(bts))

}

// authHeaderTransport is transport that wraps old tripper with auth header add.
type authHeaderTransport struct {
	T     http.RoundTripper
	token string
}

// RoundTrip is default golang http tripper interface.
func (adt *authHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("X-API-KEY", adt.token)

	return adt.T.RoundTrip(req)
}

// newAuthHeaderTransport create new authHeaderTransport by token.
func newAuthHeaderTransport(t http.RoundTripper, token string) *authHeaderTransport {
	if t == nil {
		t = http.DefaultTransport
	}

	return &authHeaderTransport{t, token}
}
