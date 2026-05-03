package authorization

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/MaxRomanov007/smart-pc-go-lib/domain/models/user"
	"golang.org/x/oauth2"
)

// Auth manages OAuth2 authentication and token lifecycle.
// It provides thread-safe access to tokens and handles automatic token refresh.
type Auth struct {
	cfg      *Config
	token    *oauth2.Token
	tokenMux sync.Mutex
}

// Load creates an Auth instance using a previously saved token.
// The token is loaded via the LoadToken function in Config and refreshed if expired.
func Load(ctx context.Context, cfg *Config) (*Auth, error) {
	const op = "lib.authorization.Load"

	token, err := cfg.loadToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to load token: %w", op, err)
	}

	return &Auth{cfg: cfg, token: token}, nil
}

// Token retrieves the current access token, refreshing it if necessary.
func (a *Auth) Token(ctx context.Context) (string, error) {
	a.tokenMux.Lock()
	defer a.tokenMux.Unlock()

	return a.tokenDangerously(ctx)
}

// TryToken attempts to retrieve the current access token without blocking.
// Returns ErrTokenLocked if the token is currently locked by another goroutine.
func (a *Auth) TryToken(ctx context.Context) (string, error) {
	if !a.tokenMux.TryLock() {
		return "", ErrTokenLocked
	}
	defer a.tokenMux.Unlock()

	return a.tokenDangerously(ctx)
}

func (a *Auth) tokenDangerously(ctx context.Context) (string, error) {
	const op = "lib.authorization.tokenDangerously"

	if a.token == nil {
		return "", ErrNoToken
	}

	if a.token.Valid() {
		return a.token.AccessToken, nil
	}

	token, err := a.cfg.refreshToken(ctx, a.token)
	if err != nil {
		return "", fmt.Errorf("%s: failed to refresh token: %w", op, err)
	}

	a.token = token
	return a.token.AccessToken, nil
}

func (a *Auth) FetchUserInfo(ctx context.Context) (*user.Info, error) {
	const op = "lib.authorization.FetchUserInfo"

	client := a.cfg.Oauth2Config.Client(ctx, a.token)

	resp, err := client.Get(a.cfg.UserInfoURL)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get user info: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: userinfo request failed, status: %s", op, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read body: %w", op, err)
	}

	info := new(user.Info)
	if err := json.Unmarshal(body, info); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal user info: %w", op, err)
	}

	return info, nil
}
