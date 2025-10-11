package commands

import (
	"context"
	"hermes-relay/internal/utils/dispatch"
)

const PingCommand = "ping"
const PongEvent = "pong"

var AddPingPongRoute = dispatch.LimitOnAction(PingCommand, addPingHandler)

type PingPayload struct {
	ResponseId string `json:"identifier"`
}

func addPingHandler(ctx context.Context, action *dispatch.Message, publisher dispatch.PublishFunc) (*dispatch.Message, error) {
	var payload PingPayload
	if err := dispatch.UnmarshalPayload(action, &payload); err != nil {
		return nil, err
	}

	return dispatch.MakeSystemEvent(
		PongEvent,
		payload,
		action,
	), nil
}
