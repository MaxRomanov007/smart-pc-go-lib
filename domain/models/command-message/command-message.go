package commandMessage

import (
	"encoding/json"
	"fmt"

	mqttMessage "github.com/MaxRomanov007/smart-pc-go-lib/domain/models/mqtt-message"
)

type Data struct {
	Command   string          `json:"command"`
	Parameter json.RawMessage `json:"parameter"`
}

type Message mqttMessage.Message[Data]

func Parameter[T any](msg *Message) (T, error) {
	const op = "models.command-message.Parameter"

	var result T
	if err := json.Unmarshal(msg.Data.Parameter, &result); err != nil {
		return result, fmt.Errorf("%s: failed to marshal payload: %w", op, err)
	}

	return result, nil
}
