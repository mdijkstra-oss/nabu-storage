package test_helpers

import (
	"hermes-relay/internal/cqrs/commands"
)

type RouterTestCase struct {
	Name              string
	InputMessage      *commands.AnyMessage
	ExpectedReturn    *commands.AnyMessage
	ExpectedPublished []*commands.AnyMessage
	ExpectError       bool
}
