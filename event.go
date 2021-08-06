package motto

type Event interface {
	Name() string
}

type Listener func(payload interface{})

type BaseEvent struct {
	name string
}

func NewEvent(name string) Event {
	return &BaseEvent{
		name: name,
	}
}

func (evt *BaseEvent) Name() string {
	return evt.name
}

type EventBus struct {
	listeners map[Event][]Listener
}

func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make(map[Event][]Listener),
	}
}

func (bus *EventBus) On(event Event, listener Listener) {
	listeners, ok := bus.listeners[event]

	if !ok {
		bus.listeners[event] = []Listener{listener}
	} else {
		bus.listeners[event] = append(listeners, listener)
	}
}

func (bus *EventBus) Fire(event Event, payload interface{}) {
	listeners, ok := bus.listeners[event]

	if ok {
		for _, listener := range listenrs {
			listener(payload)
		}
	}
}
