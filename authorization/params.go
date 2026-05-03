package authorization

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
)

const randomBytesSize = 32

type pkceAuthParams struct {
	state     string
	verifier  string
	challenge string
}

func generatePKCEParams() (*pkceAuthParams, error) {
	const op = "lib.authorization.generatePKCEParams"

	state, err := generateRandomString()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to generate state: %w", op, err)
	}

	verifier, err := generateRandomString()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to generate verifier: %w", op, err)
	}

	return &pkceAuthParams{
		state:     state,
		verifier:  verifier,
		challenge: generateChallenge(verifier),
	}, nil
}

func generateRandomString() (string, error) {
	const op = "lib.authorization.generateRandomString"

	data := make([]byte, randomBytesSize)
	if _, err := rand.Read(data); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func generateChallenge(verifier string) string {
	h := sha256.New()
	h.Write([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// getFreePort остался для обратной совместимости — может пригодиться вызывающему коду.
func getFreePort() (int, error) {
	const op = "lib.authorization.getFreePort"

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, fmt.Errorf("%s: failed to resolve address: %w", op, err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to listen: %w", op, err)
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}
