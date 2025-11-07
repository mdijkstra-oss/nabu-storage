package dispatch

import (
	"errors"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/lib/utils"
	"sync"
	"testing"
)

func TestPublisher(t *testing.T) {
	testMsg := commands.ToAny(commands.NewCommand[any, any]("TestAction", nil, "TestEntity", "test-1", nil))
	testEvent := commands.ToAny(commands.NewDomainEvent[any, any]("TestEvent", nil, "TestEntity", "test-1", nil))
	invalidMsg := &commands.AnyMessage{Type: "", Action: ""}

	emptyRouter := func(msg *commands.AnyMessage, _ PublishFunc) (*commands.AnyMessage, error) {
		return nil, nil
	}

	returnEventOnCommand := func(event *commands.AnyMessage) CommandRouter {
		return func(msg *commands.AnyMessage, _ PublishFunc) (*commands.AnyMessage, error) {
			if msg.Type == commands.Command {
				return event, nil
			}
			return nil, nil
		}
	}

	returnErrorOnCommand := func(err error) CommandRouter {
		return func(msg *commands.AnyMessage, _ PublishFunc) (*commands.AnyMessage, error) {
			if msg.Type == commands.Command {
				return nil, err
			}
			return nil, nil
		}
	}

	tests := []PublisherTestCase{
		{
			Name:            "Publish with no subscribers returns nil",
			Subscribers:     []CommandRouter{},
			Input:           testMsg,
			ExpectErr:       "",
			ExpectEvent:     nil,
			ExpectPublished: []*commands.AnyMessage{testMsg},
		},
		{
			Name:            "Publish with single subscriber that returns nil",
			Subscribers:     []CommandRouter{emptyRouter},
			Input:           testMsg,
			ExpectErr:       "",
			ExpectEvent:     nil,
			ExpectPublished: []*commands.AnyMessage{testMsg},
		},
		{
			Name:            "Publish with single subscriber that returns event",
			Subscribers:     []CommandRouter{returnEventOnCommand(testEvent)},
			Input:           testMsg,
			ExpectErr:       "",
			ExpectEvent:     testEvent,
			ExpectPublished: []*commands.AnyMessage{testMsg, testEvent},
		},
		{
			Name: "Publish returns first non-nil result from multiple subscribers",
			Subscribers: []CommandRouter{
				emptyRouter,
				returnEventOnCommand(commands.ToAny(commands.NewDomainEvent[any, any]("FirstEvent", nil, "TestEntity", "test-1", nil))),
				returnEventOnCommand(commands.ToAny(commands.NewDomainEvent[any, any]("SecondEvent", nil, "TestEntity", "test-1", nil))),
			},
			Input:       testMsg,
			ExpectErr:   "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[any, any]("FirstEvent", nil, "TestEntity", "test-1", nil)),
			ExpectPublished: []*commands.AnyMessage{
				testMsg,
				commands.ToAny(commands.NewDomainEvent[any, any]("FirstEvent", nil, "TestEntity", "test-1", nil)),
				commands.ToAny(commands.NewDomainEvent[any, any]("SecondEvent", nil, "TestEntity", "test-1", nil)),
			},
		},
		{
			Name: "Publish calls all subscribers even after getting result",
			Subscribers: []CommandRouter{
				returnEventOnCommand(testEvent),
				returnEventOnCommand(commands.ToAny(commands.NewDomainEvent[any, any]("SecondEvent", nil, "TestEntity", "test-1", nil))),
			},
			Input:       testMsg,
			ExpectErr:   "",
			ExpectEvent: testEvent,
			ExpectPublished: []*commands.AnyMessage{
				testMsg,
				testEvent,
				commands.ToAny(commands.NewDomainEvent[any, any]("SecondEvent", nil, "TestEntity", "test-1", nil)),
			},
		},
		{
			Name: "Publish stops and returns error from subscriber",
			Subscribers: []CommandRouter{
				emptyRouter,
				returnErrorOnCommand(errors.New("subscriber error")),
				returnEventOnCommand(testEvent),
			},
			Input:     testMsg,
			ExpectErr: "subscriber error",
		},
		{
			Name:        "Publish validates message before calling subscribers",
			Subscribers: []CommandRouter{},
			Input:       invalidMsg,
			ExpectErr:   "validation failed: Action is required",
		},
	}

	RunPublisherTests(t, tests)
}

func TestUnsubscribe(t *testing.T) {
	testMsg := commands.ToAny(commands.NewCommand[any, any]("TestAction", nil, "TestEntity", "test-1", nil))

	t.Run("Unsubscribe only removes specific subscriber", func(t *testing.T) {
		publisher := NewInMemoryPublisher()

		var calls1, calls2 int
		unsubscribe1 := publisher.Subscribe(func(msg *commands.AnyMessage, _ PublishFunc) (*commands.AnyMessage, error) {
			calls1++
			return nil, nil
		})
		publisher.Subscribe(func(msg *commands.AnyMessage, _ PublishFunc) (*commands.AnyMessage, error) {
			calls2++
			return nil, nil
		})

		utils.Must(publisher.Publish(testMsg))
		if calls1 != 1 || calls2 != 1 {
			t.Fatalf("expected both subscribers called once, got calls1=%d, calls2=%d", calls1, calls2)
		}

		unsubscribe1()

		utils.Must(publisher.Publish(testMsg))
		if calls1 != 1 {
			t.Fatalf("expected subscriber1 not called after unsubscribe, got %d calls", calls1)
		}
		if calls2 != 2 {
			t.Fatalf("expected subscriber2 still called, got %d calls", calls2)
		}
	})

	t.Run("Multiple unsubscribes are safe", func(t *testing.T) {
		publisher := NewInMemoryPublisher()

		unsubscribe := publisher.Subscribe(func(msg *commands.AnyMessage, _ PublishFunc) (*commands.AnyMessage, error) {
			return nil, nil
		})

		unsubscribe()
		unsubscribe()
		unsubscribe()

		_, err := publisher.Publish(testMsg)
		if err != nil {
			t.Fatalf("expected no error after multiple unsubscribes, got %v", err)
		}
	})
}

func runConcurrent(count int, fn func()) {
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn()
		}()
	}
	wg.Wait()
}

func TestConcurrency(t *testing.T) {
	testMsg := commands.ToAny(commands.NewCommand[any, any]("TestAction", nil, "TestEntity", "test-1", nil))

	t.Run("Concurrent subscribe", func(t *testing.T) {
		publisher := NewInMemoryPublisher()
		subscriberCount := 100

		runConcurrent(subscriberCount, func() {
			publisher.Subscribe(func(msg *commands.AnyMessage, _ PublishFunc) (*commands.AnyMessage, error) {
				return nil, nil
			})
		})

		if len(publisher.subscribers) != subscriberCount {
			t.Fatalf("expected %d subscribers, got %d", subscriberCount, len(publisher.subscribers))
		}
	})

	t.Run("Concurrent publish", func(t *testing.T) {
		publisher := NewInMemoryPublisher()

		var calls int
		var mu sync.Mutex
		publisher.Subscribe(func(msg *commands.AnyMessage, _ PublishFunc) (*commands.AnyMessage, error) {
			mu.Lock()
			calls++
			mu.Unlock()
			return nil, nil
		})

		publishCount := 100
		runConcurrent(publishCount, func() {
			utils.Must(publisher.Publish(testMsg))
		})

		if calls != publishCount {
			t.Fatalf("expected %d calls, got %d", publishCount, calls)
		}
	})

	t.Run("Concurrent subscribe and unsubscribe", func(t *testing.T) {
		publisher := NewInMemoryPublisher()
		operationCount := 50

		runConcurrent(operationCount, func() {
			unsubscribe := publisher.Subscribe(func(msg *commands.AnyMessage, _ PublishFunc) (*commands.AnyMessage, error) {
				return nil, nil
			})
			unsubscribe()
		})

		if len(publisher.subscribers) != 0 {
			t.Fatalf("expected 0 subscribers after all unsubscribed, got %d", len(publisher.subscribers))
		}
	})
}
