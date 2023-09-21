package hsystem

import (
	"sync"
	"syscall/js"

	"github.com/qbart/hashira/hjs"
)

type Event struct {
	Type string
	Data hjs.Object
}

type Commands struct {
	sync.RWMutex
	Events []*Event
}

func (c *Commands) AddEvent(this js.Value, args []js.Value) any {
	c.Lock()
	c.Events = append(c.Events, &Event{
		Type: args[0].String(),
		Data: hjs.Object(args[1]),
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
