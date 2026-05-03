package authorization

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/oauth2"
)

type (
	LoadTokenFunc func(context.Context) (*oauth2.Token, error)
	SaveTokenFunc func(context.Context, *oauth2.Token) error
)

// Config holds configuration for OAuth2 authorization flow.
type Config struct {
	Oauth2Config *oauth2.Config
	LoadToken    LoadTokenFunc
	SaveToken    SaveTokenFunc
	UserInfoURL  string
}

// AuthFlow holds everything needed to complete an in-progress OAuth2 PKCE flow.
// Obtain it via PrepareAuthFlow, then pass the URL to the user and call
// Finalize when the OAuth2 provider redirects back with a code.
type AuthFlow struct {
	// URL is the authorization URL the user must open in a browser.
	URL string
	// state is kept private — Finalize validates it internally.
	state    string
	verifier string
	cfg      *Config
}

// PrepareAuthFlow generates PKCE parameters and returns an AuthFlow.
// redirectURL must be the full URL of your /auth/callback route,
// e.g. "http://localhost:8080/auth/callback".
func (cfg *Config) PrepareAuthFlow(redirectURL string) (*AuthFlow, error) {
	const op = "lib.authorization.PrepareAuthFlow"

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("%s: invalid config: %w", op, err)
	}

	params, err := generatePKCEParams()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to generate PKCE params: %w", op, err)
	}

	cfg.Oauth2Config.RedirectURL = redirectURL
	url := cfg.Oauth2Config.AuthCodeURL(
		params.state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", params.challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	return &AuthFlow{
		URL:      url,
		state:    params.state,
		verifier: params.verifier,
		cfg:      cfg,
	}, nil
}

// Finalize validates the state returned by the OAuth2 provider, exchanges
// the authorization code for a token, saves it, and returns a ready Auth.
func (f *AuthFlow) Finalize(ctx context.Context, state, code string) (*Auth, error) {
	const op = "lib.authorization.AuthFlow.Finalize"

	if state != f.state {
		return nil, fmt.Errorf("%s: state mismatch", op)
	}

	token, err := f.cfg.Oauth2Config.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("code_verifier", f.verifier),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to exchange code: %w", op, err)
	}

	if err := f.cfg.saveTokenIfNeeded(ctx, token); err != nil {
		return nil, fmt.Errorf("%s: failed to save token: %w", op, err)
	}

	return &Auth{cfg: f.cfg, token: token}, nil
}

// loadToken loads a saved token and refreshes it if expired.
func (cfg *Config) loadToken(ctx context.Context) (*oauth2.Token, error) {
	const op = "lib.authorization.config.loadToken"

	if cfg.LoadToken == nil {
		return nil, fmt.Errorf("%s: LoadToken is not defined", op)
	}

	token, err := cfg.LoadToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !token.Valid() {
		token, err = cfg.Oauth2Config.TokenSource(ctx, token).Token()
		if err != nil {
			return nil, fmt.Errorf("%s: failed to refresh token: %w", op, err)
		}

		if err := cfg.saveTokenIfNeeded(ctx, token); err != nil {
			return nil, fmt.Errorf("%s: failed to save refreshed token: %w", op, err)
		}
	}

	return token, nil
}

func (cfg *Config) refreshToken(ctx context.Context, t *oauth2.Token) (*oauth2.Token, error) {
	const op = "lib.authorization.config.refreshToken"

	token, err := cfg.Oauth2Config.TokenSource(ctx, t).Token()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to refresh: %w", op, err)
	}

	if err := cfg.saveTokenIfNeeded(ctx, token); err != nil {
		return nil, fmt.Errorf("%s: failed to save: %w", op, err)
	}

	return token, nil
}

func (cfg *Config) saveTokenIfNeeded(ctx context.Context, token *oauth2.Token) error {
	const op = "lib.authorization.config.saveTokenIfNeeded"

	if cfg.SaveToken == nil {
		return nil
	}

	if err := cfg.SaveToken(ctx, token); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (cfg *Config) validate() error {
	var errs []error

	if cfg.Oauth2Config.ClientID == "" {
		errs = append(errs, errors.New("missing client id"))
	}
	if cfg.Oauth2Config.Endpoint.AuthURL == "" {
		errs = append(errs, errors.New("missing auth url"))
	}
	if cfg.Oauth2Config.Endpoint.TokenURL == "" {
		errs = append(errs, errors.New("missing token url"))
	}

	return errors.Join(errs...)
}
