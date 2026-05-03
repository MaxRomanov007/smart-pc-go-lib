package authorization

import "errors"

var (
	// ErrNoToken indicates that no token is available for authentication.
	ErrNoToken = errors.New("no token")
	// ErrTokenLocked indicates that the token is currently locked by another goroutine.
	ErrTokenLocked = errors.New("token is locked")
)
