package hsystem

import (
	"sync"
	"syscall/js"
)

type Event struct {
	Type    string
	Payload []byte
}

type Commands struct {
	sync.RWMutex
	Events []*Event
}

func (c *Commands) AddEvent(this js.Value, args []js.Value) any {
	c.Lock()
	id := args[0].String()
	data := args[1].String()
	c.Events = append(c.Events, &Event{
		Type:    id,
		Payload: []byte(data),
	})
	defer c.Unlock()

	return js.Null()
}

func (c *Commands) HasEvents() bool {
	c.RLock()
	defer c.RUnlock()
	return len(c.Events) > 0
}

func (c *Commands) PeekEvent() *Event {
	c.Lock()
	defer c.Unlock()
	event := c.Events[0]
	c.Events = c.Events[1:]
	return event
}
