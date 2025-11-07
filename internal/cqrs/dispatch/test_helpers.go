package dispatch

import (
	"hermes-relay/internal/cqrs/commands"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

type PublisherTestCase struct {
	Name            string
	Subscribers     []CommandRouter
	Input           *commands.AnyMessage
	ExpectErr       string
	ExpectEvent     *commands.AnyMessage
	ExpectPublished []*commands.AnyMessage
}

func RunPublisherTests(t *testing.T, tests []PublisherTestCase) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			publisher := NewInMemoryPublisher()

			var published []*commands.AnyMessage
			publisher.Subscribe(func(msg *commands.AnyMessage, _ PublishFunc) (*commands.AnyMessage, error) {
				published = append(published, msg)
				return nil, nil
			})

			for _, sub := range tt.Subscribers {
				publisher.Subscribe(sub)
			}

			result, err := publisher.Publish(tt.Input)

			th.AssertError(t, err, tt.ExpectErr, "error")
			if tt.ExpectErr == "" {
				th.AssertMessage(t, result, tt.ExpectEvent, "event")

				if tt.ExpectPublished != nil {
					if len(published) != len(tt.ExpectPublished) {
						t.Fatalf("published events count: expected %d, got %d", len(tt.ExpectPublished), len(published))
					}
					for i, expected := range tt.ExpectPublished {
						th.AssertMessage(t, published[i], expected, "published["+string(rune(i))+"]")
					}
				} else if len(published) > 0 {
					t.Fatalf("expected no published events, but got %d", len(published))
				}
			}
		})
	}
}
