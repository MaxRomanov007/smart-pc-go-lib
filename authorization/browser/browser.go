package browser

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/MaxRomanov007/smart-pc-go-lib/authorization"
	goBrowser "github.com/MaxRomanov007/smart-pc-go-lib/cross-platform/browser"
)

// Authorize выполняет полный OAuth2 PKCE flow с открытием браузера.
func Authorize(
	ctx context.Context,
	cfg *authorization.Config,
	cbCfg CallbackConfig,
) (*authorization.Auth, error) {
	const op = "authorization.browser.Authorize"

	port, err := getFreePort()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get free port: %w", op, err)
	}

	redirectURL := fmt.Sprintf("http://%s:%d/callback", cbCfg.Host, port)

	flow, err := cfg.PrepareAuthFlow(redirectURL)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to prepare auth flow: %w", op, err)
	}

	if err := goBrowser.Open(flow.URL); err != nil {
		fmt.Printf("Open this link to authorize:\n%s\n", flow.URL)
	}

	state, code, err := waitForCode(ctx, cbCfg, port)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get callback code: %w", op, err)
	}

	auth, err := flow.Finalize(ctx, state, code)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to finalize: %w", op, err)
	}

	return auth, nil
}

// waitForCode поднимает временный HTTP-сервер и ждёт OAuth2 callback.
// Возвращает state и code из query-параметров редиректа.
func waitForCode(
	ctx context.Context,
	cbCfg CallbackConfig,
	port int,
) (state, code string, err error) {
	const op = "authorization.browser.waitForCode"

	timeoutCtx, cancel := context.WithTimeout(ctx, cbCfg.TTL)
	defer cancel()

	codeChan := make(chan [2]string, 1) // [state, code]
	errChan := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		s := r.URL.Query().Get("state")
		c := r.URL.Query().Get("code")

		if s == "" || c == "" {
			http.Error(w, "missing state or code", http.StatusBadRequest)
			errChan <- fmt.Errorf("%s: missing state or code in callback", op)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, "Authorization succeeded")
		codeChan <- [2]string{s, c}
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cbCfg.Host, port),
		Handler:      mux,
		ReadTimeout:  cbCfg.ReadTimeout,
		WriteTimeout: cbCfg.WriteTimeout,
		IdleTimeout:  cbCfg.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("%s: callback server error: %w", op, err)
		}
	}()

	select {
	case pair := <-codeChan:
		_ = srv.Shutdown(context.Background())
		return pair[0], pair[1], nil
	case err := <-errChan:
		_ = srv.Shutdown(context.Background())
		return "", "", err
	case <-timeoutCtx.Done():
		_ = srv.Shutdown(context.Background())
		return "", "", fmt.Errorf("%s: timeout waiting for callback", op)
	}
}

// getFreePort finds an available TCP port on localhost
// Uses net.ListenTCP on port 0 to let the OS assign a free port
// Returns the assigned port number or an error if no port is available
func getFreePort() (int, error) {
	const op = "lib.authorization.config.getFreePort"

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, fmt.Errorf("%s: failed to resolve tcp address: %w", op, err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to listen tcp: %w", op, err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
