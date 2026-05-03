// Package apiclient provides a generic HTTP client that works with
// github.com/MaxRomanov007/smart-pc-go-lib/api/response envelopes.
//
// It is intentionally decoupled from the authorization package so it can be
// used anywhere a Bearer token is available, not only in OAuth2 flows.
package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/MaxRomanov007/smart-pc-go-lib/api/response"
)

// TokenProvider is the only dependency the Client has on authentication.
// *authorization.Auth satisfies this interface automatically.
type TokenProvider interface {
	Token(ctx context.Context) (string, error)
}

// Client is a thin wrapper around *http.Client that automatically attaches
// a Bearer token to every request and decodes response.Response envelopes.
type Client struct {
	http  *http.Client
	token TokenProvider
	// UID is populated from the OAuth2 UserInfo endpoint by NewWithUserInfo.
	// It is empty when the client is created via New.
	UID string
}

// New creates a Client. httpClient may be nil — http.DefaultClient is used in that case.
func New(httpClient *http.Client, token TokenProvider) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{http: httpClient, token: token}
}

// NewWithUID creates a Client with a pre-populated UID field.
func NewWithUID(httpClient *http.Client, token TokenProvider, uid string) *Client {
	c := New(httpClient, token)
	c.UID = uid
	return c
}

// NewRequest builds an *http.Request with the Authorization header already set.
func (c *Client) NewRequest(
	ctx context.Context,
	method, url string,
	body io.Reader,
) (*http.Request, error) {
	const op = "apiclient.NewRequest"

	token, err := c.token.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get token: %w", op, err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to build request: %w", op, err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	return req, nil
}

// Do executes req and decodes the response envelope into response.Response[T].
func Do[T any](c *Client, req *http.Request) (*response.Response[T], error) {
	const op = "apiclient.Do"

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: request failed: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: unexpected status %s", op, resp.Status)
	}

	result := new(response.Response[T])
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, fmt.Errorf("%s: failed to decode response: %w", op, err)
	}

	return result, nil
}

// Send is a convenience function: it builds a request with an optional JSON body,
// executes it, and decodes the response envelope — all in one call.
func Send[T any](
	ctx context.Context,
	c *Client,
	method, url string,
	body any,
) (*response.Response[T], error) {
	const op = "apiclient.Send"

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to marshal body: %w", op, err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := c.NewRequest(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	result, err := Do[T](c, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}
