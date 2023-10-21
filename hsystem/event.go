package hsystem

import (
	"encoding/json"
	"sync"
	"syscall/js"
)

type Event struct {
	Type     string
	JsonData string
}

type Commands struct {
	sync.RWMutex
	Events []*Event
}

func JsonData[T any](data string) *T {
	var parsed T
	_ = json.Unmarshal([]byte(data), &parsed)
	return &parsed
}

func (c *Commands) AddEvent(this js.Value, args []js.Value) any {
	c.Lock()
	id := args[0].String()
	data := args[1].String()
	c.Events = append(c.Events, &Event{
		Type:     id,
		JsonData: data,
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
