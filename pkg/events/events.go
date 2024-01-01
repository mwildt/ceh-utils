package events

import (
	"encoding/json"
	"fmt"
)

type eventType string

type subscription func(Event) error

type subscriptions []subscription

type Event struct {
	Type       eventType
	Payload    []byte
	ContenType string
}

type eventBus struct {
	subscriber map[eventType]subscriptions
}

var bus eventBus = eventBus{subscriber: make(map[eventType]subscriptions)}

func (bus *eventBus) emit(event Event) {
	// es werden erstmal die konkreten subscriber bedient
	subscriptions, ok := bus.subscriber[event.Type]
	if ok && len(subscriptions) > 0 {
		for _, sub := range subscriptions {
			sub(event)
		}
	}

	// und dann noch ein globaler subscriber
	subscriptions, ok = bus.subscriber[eventType("*")]
	if ok && len(subscriptions) > 0 {
		for _, sub := range subscriptions {
			sub(event)
		}
	}
}

func (bus *eventBus) subscribe(eType eventType, subscription subscription) error {
	if _, ok := bus.subscriber[eType]; !ok {
		bus.subscriber[eType] = make(subscriptions, 0)
	}
	bus.subscriber[eType] = append(bus.subscriber[eType], subscription)
	return nil
}

func Emit(eType string, event interface{}) error {
	if payload, err := json.Marshal(event); err != nil {
		return err
	} else {
		bus.emit(Event{Type: eventType(eType), Payload: payload, ContenType: "application/json"})
	}
	return nil
}

func Subscribe[T any](eType string, handler func(T) error) error {
	return bus.subscribe(eventType(eType), func(event Event) error {
		var payload T
		switch event.ContenType {
		case "application/json":
			{
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					return err
				}
			}
		default:
			{
				return fmt.Errorf("illegal event payload type %s for event type %s", event.ContenType, event.Type)
			}
		}
		// invoke des eigentlichen Handlers
		return handler(payload)
	})
}

func SubscribeRaw(eType string, handler func(event Event) error) error {
	return bus.subscribe(eventType(eType), func(event Event) error {
		return handler(event)
	})
}
