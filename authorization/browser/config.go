package browser

import "time"

type CallbackConfig struct {
	Host         string        // Host for callback server (e.g., "127.0.0.1")
	TTL          time.Duration // Maximum time to wait for callback
	ReadTimeout  time.Duration // HTTP server read timeout
	WriteTimeout time.Duration // HTTP server write timeout
	IdleTimeout  time.Duration // HTTP server idle timeout
}
