package authorization

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/MaxRomanov007/smart-pc-go-lib/api/response"
)

type ApiClient struct {
	client *http.Client
	auth   *Auth
	UID    string
}

func (a *Auth) NewApiClient(ctx context.Context, client *http.Client) (*ApiClient, error) {
	const op = "authorization.NewApiClient"

	userinfo, err := a.FetchUserInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to fetch user info: %w", op, err)
	}

	return &ApiClient{
		client: client,
		auth:   a,
		UID:    userinfo.Sub,
	}, nil
}

func (ac *ApiClient) NewRequest(
	ctx context.Context,
	method string,
	url string,
	body io.Reader,
) (*http.Request, error) {
	const op = "authorization.api-client.NewRequest"

	token, err := ac.auth.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get auth token: %w", op, err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create request: %w", op, err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	return req, nil
}

func DoRequest[T any](ac *ApiClient, req *http.Request) (*response.Response[T], error) {
	const op = "authorization.DoRequest"

	resp, err := ac.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to do request: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: bad status: %s", op, resp.Status)
	}

	result := new(response.Response[T])
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, fmt.Errorf("%s: failed to decode response: %w", op, err)
	}

	return result, nil
}

func DoNewRequest[T any](
	ctx context.Context,
	ac *ApiClient,
	method string,
	url string,
	body io.Reader,
) (*response.Response[T], error) {
	const op = "authorization.DoNewRequest"

	req, err := ac.NewRequest(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create request: %w", op, err)
	}

	resp, err := DoRequest[T](ac, req)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to do request: %w", op, err)
	}

	return resp, nil
}
