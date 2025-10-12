package pingpong

import (
	"context"
	commands2 "hermes-relay/internal/commands"
)

const PingCommand = "ping"
const PongEvent = "pong"

var AddPingPongRoute = commands2.LimitOnAction(PingCommand, addPingHandler)

type PingPayload struct {
	ResponseId string `json:"identifier"`
}

func addPingHandler(ctx context.Context, action *commands2.Message, publisher commands2.PublishFunc) (*commands2.Message, error) {
	var payload PingPayload
	if err := commands2.UnmarshalPayload(action, &payload); err != nil {
		return nil, err
	}

	return commands2.MakeSystemEvent(
		PongEvent,
		payload,
		action,
	), nil
}
