// Package eventbus provides a straightforward concurrent EventBus for Go 1.18+, supporting fanout, and in order, only once delivery.
package eventbus

import (
	"sync"

	"github.com/JosiahWitt/eventbus/internal/typedsyncmap"
)

// DefaultBufferSize for the subscription channels. Used when the BufferSize is not configured.
const DefaultBufferSize = 10

// Config can be passed to NewWithConfig to customize the EventBus.
type Config struct {
	// BufferSize for the subscription channels.
	// If not set or zero, it defaults to DefaultBufferSize.
	// If negative, it creates the channels with no buffer.
	BufferSize int
}

// EventBus is a straightforward concurrent EventBus for Go 1.18+, supporting fanout, and in order, only once delivery.
//
// It can be used by initializing a copy of the struct, or by calling the New or NewWithConfig functions.
type EventBus[Event any] struct {
	topics typedsyncmap.Map[string, *topic[Event]]

	rawBufferSize int
}

type topic[Event any] struct {
	key  string
	subs map[*Subscription[Event]]*Subscription[Event]
	mu   sync.RWMutex

	bus *EventBus[Event]

	isClosed bool
}

// Subscription maintains subscriptions to multiple topics.
// Events are sent to the Channel().
type Subscription[Event any] struct {
	ch chan Event
	mu sync.Mutex

	topics []*topic[Event]
	self   *Subscription[Event]
}

// New creates a new EventBus.
func New[Event any]() *EventBus[Event] {
	return NewWithConfig[Event](&Config{})
}

// NewWithConfig creates a new customized EventBus.
func NewWithConfig[Event any](config *Config) *EventBus[Event] {
	return &EventBus[Event]{
		rawBufferSize: config.BufferSize,
	}
}

// Publish sends the provided event to all of the listed topics.
// All subscriptions to those topics will be notified of the event.
func (b *EventBus[Event]) Publish(event Event, topicKeys ...string) {
	publishedSubscriptions := map[*Subscription[Event]]bool{}

	for _, topicKey := range topicKeys {
		b.publishToTopic(topicKey, event, publishedSubscriptions)
	}
}

// Subscribe creates a new subscription to the listed topics.
// All events published to any of those topics will be sent to the subscription's channel.
//
// If the same event is sent to multiple of the listed topics, the event will only be delivered once.
func (b *EventBus[Event]) Subscribe(topicKeys ...string) *Subscription[Event] {
	sub := &Subscription[Event]{
		ch: make(chan Event, b.bufferSizeOrDefault()),
	}
	sub.self = sub

	for _, topicKey := range topicKeys {
		topic := b.findOrCreateTopic(topicKey)
		topic.addSubscription(sub)

		sub.topics = append(sub.topics, topic)
	}

	return sub
}

// Unsubscribe closes the subscription to the topics.
// It also closes the subscription's channel.
func (s *Subscription[Event]) Unsubscribe() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove the subscription from the topics behind a mutex (inside t.removeSubscription)
	// before closing the channel to prevent writing to a closed channel
	for _, t := range s.topics {
		t.removeSubscription(s)
	}

	close(s.ch)
}

// Channel exposes a read only view of the subscription's channel.
// All events published to the subscribed topics will be published to this channel.
//
// If the same event is sent to multiple of the listed topics, the event will only be delivered once.
func (sub *Subscription[Event]) Channel() <-chan Event {
	return sub.ch
}

func (b *EventBus[Event]) findOrCreateTopic(topicKey string) *topic[Event] {
	t, ok := b.topics.Load(topicKey)
	if !ok {
		// If another goroutine created a topic at the same time,
		// we'd want to use it and not create a duplicate
		t, _ = b.topics.LoadOrStore(topicKey, &topic[Event]{
			key: topicKey,
			bus: b,

			subs: make(map[*Subscription[Event]]*Subscription[Event]),
		})
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	// Make sure the topic wasn't closed while we were loading it
	if t.isClosed {
		return b.findOrCreateTopic(topicKey)
	}

	return t
}

func (t *topic[Event]) addSubscription(s *Subscription[Event]) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.subs[s] = s
}

func (t *topic[Event]) removeSubscription(sub *Subscription[Event]) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.subs, sub)

	if len(t.subs) == 0 {
		t.isClosed = true
		t.bus.topics.Delete(t.key)
	}
}

func (b *EventBus[Event]) publishToTopic(topicKey string, event Event, publishedSubscriptions map[*Subscription[Event]]bool) {
	t, ok := b.topics.Load(topicKey)
	if !ok {
		return
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, sub := range t.subs {
		// If we already published to this subscription, don't publish again to guarantee only once delivery
		if _, alreadyPublished := publishedSubscriptions[sub]; alreadyPublished {
			continue
		}

		sub.ch <- event
		publishedSubscriptions[sub] = true
	}
}

func (b *EventBus[Event]) bufferSizeOrDefault() int {
	if b.rawBufferSize == 0 {
		return DefaultBufferSize
	} else if b.rawBufferSize < 0 {
		return 0
	}

	return b.rawBufferSize
}
