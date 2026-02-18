package message

import (
	"encoding/json"
	"fmt"
)

type Data struct {
	Command   string          `json:"command"`
	Parameter json.RawMessage `json:"parameter"`
}

type Message struct {
	Type string `json:"type"`
	Data Data   `json:"data"`
}

func Parameter[T any](msg *Message) (T, error) {
	const op = "lib.commands.message.Parameter"

	var result T
	if err := json.Unmarshal(msg.Data.Parameter, &result); err != nil {
		return result, fmt.Errorf("failed to marshal payload: %w", err)
	}

	return result, nil
}
