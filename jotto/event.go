package jotto

// Event represents an event in Motto
type Event interface {
	Name() string
}

// Listener is an event listener that will be called when an event happens.
type Listener func(payload ...interface{})

// BaseEvent is the built-in event representation of Motto
type BaseEvent struct {
	name string
}

// NewEvent creates a new BaseEvent.
func NewEvent(name string) *BaseEvent {
	return &BaseEvent{
		name: name,
	}
}

// Name returns the event name
func (evt *BaseEvent) Name() string {
	return evt.name
}

// EventBus is a container that stores events and their mappings to event listeners.
type EventBus struct {
	listeners map[Event][]Listener
}

// NewEventBus creates a new EventBus
func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make(map[Event][]Listener),
	}
}

// On registers a `listener` to the `event`.
func (bus *EventBus) On(event Event, listener Listener) {
	listeners, ok := bus.listeners[event]

	if !ok {
		bus.listeners[event] = []Listener{listener}
	} else {
		bus.listeners[event] = append(listeners, listener)
	}
}

// Fire emits an `event` with a `payload`.
// All listeners registered under this event will be called syncronously.
func (bus *EventBus) Fire(event Event, payload ...interface{}) {
	listeners, ok := bus.listeners[event]

	if ok {
		for _, listener := range listeners {
			listener(payload...)
		}
	}
}
