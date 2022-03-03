package eventbus_test

import (
	"sync"
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensurepkg"
	"github.com/JosiahWitt/eventbus"
)

type MyEvent struct {
	ID string
}

func TestEventBus(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("when nothing is subscribed", func(ensure ensurepkg.Ensure) {
		bus := eventbus.New[*MyEvent]()

		bus.Publish(&MyEvent{ID: "123"}, "key1", "key2")
	})

	ensure.Run("when one topic is subscribed and one event is published", func(ensure ensurepkg.Ensure) {
		ensure.Run("when created by New", func(ensure ensurepkg.Ensure) {
			bus := eventbus.New[*MyEvent]()

			sub1 := bus.Subscribe("key1")
			buf1 := bufferSubscription(sub1, 1)

			bus.Publish(&MyEvent{ID: "123"}, "key1", "key2")

			ensure(buf1.events()).Equals([]*MyEvent{{ID: "123"}})
		})

		ensure.Run("when created by initializing the struct", func(ensure ensurepkg.Ensure) {
			bus := eventbus.EventBus[*MyEvent]{}

			sub1 := bus.Subscribe("key1")
			buf1 := bufferSubscription(sub1, 1)

			bus.Publish(&MyEvent{ID: "123"}, "key1", "key2")

			ensure(buf1.events()).Equals([]*MyEvent{{ID: "123"}})
		})

		ensure.Run("when created by NewWithConfig", func(ensure ensurepkg.Ensure) {
			bus := eventbus.NewWithConfig[*MyEvent](&eventbus.Config{})

			sub1 := bus.Subscribe("key1")
			buf1 := bufferSubscription(sub1, 1)

			bus.Publish(&MyEvent{ID: "123"}, "key1", "key2")

			ensure(buf1.events()).Equals([]*MyEvent{{ID: "123"}})
		})
	})

	ensure.Run("when the buffer size is customized", func(ensure ensurepkg.Ensure) {
		bus := eventbus.NewWithConfig[*MyEvent](&eventbus.Config{
			BufferSize: 100,
		})

		sub1 := bus.Subscribe("key1")
		buf1 := bufferSubscription(sub1, 1)

		bus.Publish(&MyEvent{ID: "123"}, "key1", "key2")

		ensure(buf1.events()).Equals([]*MyEvent{{ID: "123"}})
	})

	ensure.Run("when one topic is subscribed and two events are published", func(ensure ensurepkg.Ensure) {
		bus := eventbus.New[*MyEvent]()

		sub1 := bus.Subscribe("key1")
		buf1 := bufferSubscription(sub1, 2)

		bus.Publish(&MyEvent{ID: "123"}, "key1", "key2")
		bus.Publish(&MyEvent{ID: "456"}, "key1", "key2")

		ensure(buf1.events()).Equals([]*MyEvent{{ID: "123"}, {ID: "456"}})
	})

	ensure.Run("when two topics are subscribed to different keys and two events are published", func(ensure ensurepkg.Ensure) {
		bus := eventbus.New[*MyEvent]()

		sub1 := bus.Subscribe("key1")
		buf1 := bufferSubscription(sub1, 2)

		sub2 := bus.Subscribe("key2")
		buf2 := bufferSubscription(sub2, 2)

		bus.Publish(&MyEvent{ID: "123"}, "key1", "key2")
		bus.Publish(&MyEvent{ID: "456"}, "key1", "key2")

		ensure(buf1.events()).Equals([]*MyEvent{{ID: "123"}, {ID: "456"}})
		ensure(buf2.events()).Equals([]*MyEvent{{ID: "123"}, {ID: "456"}})
	})

	ensure.Run("when two topics are subscribed to the same key and two events are published", func(ensure ensurepkg.Ensure) {
		bus := eventbus.New[*MyEvent]()

		sub1 := bus.Subscribe("key1")
		buf1 := bufferSubscription(sub1, 2)

		sub2 := bus.Subscribe("key1")
		buf2 := bufferSubscription(sub2, 2)

		bus.Publish(&MyEvent{ID: "123"}, "key1", "key2")
		bus.Publish(&MyEvent{ID: "456"}, "key1", "key2")

		ensure(buf1.events()).Equals([]*MyEvent{{ID: "123"}, {ID: "456"}})
		ensure(buf2.events()).Equals([]*MyEvent{{ID: "123"}, {ID: "456"}})
	})

	ensure.Run("only receives event once when the same event is published to duplicate topics", func(ensure ensurepkg.Ensure) {
		bus := eventbus.New[*MyEvent]()

		sub1 := bus.Subscribe("key1")
		buf1 := bufferSubscription(sub1, 2)

		bus.Publish(&MyEvent{ID: "123"}, "key1", "key2", "key1")
		bus.Publish(&MyEvent{ID: "456"}, "key1", "key2", "key1")

		ensure(buf1.events()).Equals([]*MyEvent{{ID: "123"}, {ID: "456"}})
	})

	ensure.Run("when subscribed to multiple topics", func(ensure ensurepkg.Ensure) {
		bus := eventbus.New[*MyEvent]()

		sub1 := bus.Subscribe("key1", "key2")
		buf1 := bufferSubscription(sub1, 3)

		bus.Publish(&MyEvent{ID: "123"}, "key1", "key2", "key3") // Shows that events are only delivered once
		bus.Publish(&MyEvent{ID: "456"}, "key3")
		bus.Publish(&MyEvent{ID: "789"}, "key1")
		bus.Publish(&MyEvent{ID: "42"}, "key2")

		ensure(buf1.events()).Equals([]*MyEvent{{ID: "123"}, {ID: "789"}, {ID: "42"}})
	})

	ensure.Run("when two topics are subscribed to the different key and two non-overlapping events are published", func(ensure ensurepkg.Ensure) {
		bus := eventbus.New[*MyEvent]()

		sub1 := bus.Subscribe("key1")
		buf1 := bufferSubscription(sub1, 1)

		sub2 := bus.Subscribe("key2")
		buf2 := bufferSubscription(sub2, 1)

		bus.Publish(&MyEvent{ID: "123"}, "key1")
		bus.Publish(&MyEvent{ID: "456"}, "key2")

		ensure(buf1.events()).Equals([]*MyEvent{{ID: "123"}})
		ensure(buf2.events()).Equals([]*MyEvent{{ID: "456"}})
	})

	ensure.Run("when two topics are subscribed to the different key and three variously-overlapping events are published", func(ensure ensurepkg.Ensure) {
		bus := eventbus.New[*MyEvent]()

		sub1 := bus.Subscribe("key1")
		buf1 := bufferSubscription(sub1, 2)

		sub2 := bus.Subscribe("key2")
		buf2 := bufferSubscription(sub2, 2)

		bus.Publish(&MyEvent{ID: "123"}, "key1")
		bus.Publish(&MyEvent{ID: "456"}, "key2")
		bus.Publish(&MyEvent{ID: "789"}, "key2", "key1")

		ensure(buf1.events()).Equals([]*MyEvent{{ID: "123"}, {ID: "789"}})
		ensure(buf2.events()).Equals([]*MyEvent{{ID: "456"}, {ID: "789"}})
	})

	ensure.Run("when one subscription is unsubscribed part way through", func(ensure ensurepkg.Ensure) {
		bus := eventbus.New[*MyEvent]()

		sub1 := bus.Subscribe("key1")
		buf1 := bufferSubscription(sub1, 2)

		sub2 := bus.Subscribe("key1")
		buf2 := bufferSubscription(sub2, 3)

		bus.Publish(&MyEvent{ID: "123"}, "key1")
		bus.Publish(&MyEvent{ID: "456"}, "key1")
		sub1.Unsubscribe()
		bus.Publish(&MyEvent{ID: "789"}, "key1")

		ensure(buf1.events()).Equals([]*MyEvent{{ID: "123"}, {ID: "456"}})
		ensure(buf2.events()).Equals([]*MyEvent{{ID: "123"}, {ID: "456"}, {ID: "789"}})

		_, sub1IsOpen := <-sub1.Channel()
		ensure(sub1IsOpen).IsFalse()
	})

	ensure.Run("when all subscriptions are unsubscribed on one topic and then a new one is subscribed", func(ensure ensurepkg.Ensure) {
		bus := eventbus.New[*MyEvent]()

		sub1 := bus.Subscribe("key1")
		buf1 := bufferSubscription(sub1, 2)

		sub2 := bus.Subscribe("key1")
		buf2 := bufferSubscription(sub2, 3)

		sub3 := bus.Subscribe("key2")
		buf3 := bufferSubscription(sub3, 1)

		bus.Publish(&MyEvent{ID: "123"}, "key1")
		bus.Publish(&MyEvent{ID: "456"}, "key1", "key2")
		sub1.Unsubscribe()
		bus.Publish(&MyEvent{ID: "789"}, "key1")
		sub2.Unsubscribe()

		ensure(buf1.events()).Equals([]*MyEvent{{ID: "123"}, {ID: "456"}})
		ensure(buf2.events()).Equals([]*MyEvent{{ID: "123"}, {ID: "456"}, {ID: "789"}})
		ensure(buf3.events()).Equals([]*MyEvent{{ID: "456"}})

		_, sub1IsOpen := <-sub1.Channel()
		ensure(sub1IsOpen).IsFalse()

		_, sub2IsOpen := <-sub2.Channel()
		ensure(sub2IsOpen).IsFalse()

		sub4 := bus.Subscribe("key1")
		buf4 := bufferSubscription(sub4, 1)

		bus.Publish(&MyEvent{ID: "42"}, "key1")

		ensure(buf4.events()).Equals([]*MyEvent{{ID: "42"}})
	})

	ensure.Run("concurrent publishing", func(ensure ensurepkg.Ensure) {
		bus := eventbus.New[string]()

		sub1 := bus.Subscribe("key1")
		buf1 := bufferSubscription(sub1, 3)

		sub2 := bus.Subscribe("key1")
		buf2 := bufferSubscription(sub2, 3)

		sub3 := bus.Subscribe("key2")
		buf3 := bufferSubscription(sub3, 2)

		go func() {
			bus.Publish("1", "key1")
		}()

		go func() {
			bus.Publish("2", "key2")
		}()

		go func() {
			bus.Publish("3", "key1")
		}()

		go func() {
			bus.Publish("4", "key2")
		}()

		go func() {
			bus.Publish("5", "key1")
		}()

		ensure(len(buf1.events())).Equals(3)
		ensure(buf1.events()).Contains("1")
		ensure(buf1.events()).Contains("3")
		ensure(buf1.events()).Contains("5")

		ensure(len(buf2.events())).Equals(3)
		ensure(buf2.events()).Contains("1")
		ensure(buf2.events()).Contains("3")
		ensure(buf2.events()).Contains("5")

		ensure(len(buf3.events())).Equals(2)
		ensure(buf3.events()).Contains("2")
		ensure(buf3.events()).Contains("4")
	})

	ensure.Run("concurrent subscriptions, publishing, and unsubscriptions", func(ensure ensurepkg.Ensure) {
		// This test is largely designed to help surface any race condition issues, thus the results are not checked

		bus := eventbus.New[*MyEvent]()

		// Thoroughly exercise the concurrent code
		const numParallel = 1000

		// Make sure the subscription and unsubscription can happen in separate goroutines
		for i := 0; i < numParallel; i++ {
			var (
				sub *eventbus.Subscription[*MyEvent]
				mu  sync.Mutex
			)

			mu.Lock() // Force the subscription to happen before unsubscribe is called
			go func() {
				defer mu.Unlock()
				sub = bus.Subscribe("key1")
			}()

			go func() {
				mu.Lock()
				defer mu.Unlock()
				sub.Unsubscribe()
			}()
		}

		for i := 0; i < numParallel; i++ {
			go func() {
				bus.Publish(&MyEvent{ID: "1"}, "key1")
			}()
		}

		for i := 0; i < numParallel; i++ {
			go func() {
				bus.Subscribe("key1").Unsubscribe()
			}()
		}

		for i := 0; i < numParallel; i++ {
			go func() {
				bus.Subscribe("key2").Unsubscribe()
			}()
		}

		for i := 0; i < numParallel; i++ {
			go func() {
				bus.Publish(&MyEvent{ID: "2"}, "key2")
			}()
		}

		for i := 0; i < numParallel; i++ {
			go func() {
				bus.Publish(&MyEvent{ID: "3"}, "key1")
			}()
		}

		for i := 0; i < numParallel; i++ {
			go func() {
				bus.Subscribe("key3").Unsubscribe()
			}()
		}

		for i := 0; i < numParallel; i++ {
			go func() {
				bus.Publish(&MyEvent{ID: "4"}, "key2")
			}()
		}

		for i := 0; i < numParallel; i++ {
			go func() {
				bus.Publish(&MyEvent{ID: "5"}, "key1")
			}()
		}
	})
}

func TestConfig(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("default configuration via New", func(ensure ensurepkg.Ensure) {
		bus := eventbus.New[string]()

		sub := bus.Subscribe()
		ensure(cap(sub.Channel())).Equals(eventbus.DefaultBufferSize)
	})

	ensure.Run("default configuration via initializing struct", func(ensure ensurepkg.Ensure) {
		bus := eventbus.EventBus[string]{}

		sub := bus.Subscribe()
		ensure(cap(sub.Channel())).Equals(eventbus.DefaultBufferSize)
	})

	ensure.Run("default configuration via NewWithConfig", func(ensure ensurepkg.Ensure) {
		bus := eventbus.NewWithConfig[string](&eventbus.Config{})

		sub := bus.Subscribe()
		ensure(cap(sub.Channel())).Equals(eventbus.DefaultBufferSize)
	})

	ensure.Run("force a zero buffer size", func(ensure ensurepkg.Ensure) {
		bus := eventbus.NewWithConfig[string](&eventbus.Config{
			BufferSize: -1, // Sets the buffer to zero
		})

		sub := bus.Subscribe()
		ensure(cap(sub.Channel())).Equals(0)
	})

	ensure.Run("override buffer size", func(ensure ensurepkg.Ensure) {
		bus := eventbus.NewWithConfig[string](&eventbus.Config{
			BufferSize: 100,
		})

		sub := bus.Subscribe()
		ensure(cap(sub.Channel())).Equals(100)
	})
}

type buffer[E any] struct {
	internalEvents []E
	mu             sync.Mutex
}

func bufferSubscription[E any](sub *eventbus.Subscription[E], expectedCount int) *buffer[E] {
	buf := &buffer[E]{}

	// Lock until all expected events are read
	buf.mu.Lock()
	go func() {
		defer buf.mu.Unlock()

		for i := 0; i < expectedCount; i++ {
			buf.internalEvents = append(buf.internalEvents, <-sub.Channel())
		}
	}()

	return buf
}

func (buf *buffer[E]) events() []E {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	return buf.internalEvents
}
