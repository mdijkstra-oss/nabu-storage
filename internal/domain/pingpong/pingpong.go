package pingpong

import (
	"context"
	"hermes-relay/internal/commands"
)

const PingCommand = "ping"
const PongEvent = "pong"

var AddPingPongRoute = commands.LimitOnAction(PingCommand, addPingHandler)

type PingPayload struct {
	ResponseId string `json:"identifier"`
}

func addPingHandler(ctx context.Context, action *commands.Message, publisher commands.PublishFunc) (*commands.Message, error) {
	var payload PingPayload
	if err := commands.UnmarshalPayload(action, &payload); err != nil {
		return nil, err
	}

	return commands.MakeSystemEvent(
		PongEvent,
		payload,
		action,
	), nil
}
