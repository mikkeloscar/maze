package checker

import (
	"context"
)

const key = "checkerState"

// Setter defines a context that enables setting values.
type Setter interface {
	Set(string, interface{})
}

// FromContext returns the Store associated with this context.
func FromContext(c context.Context) *State {
	return c.Value(key).(*State)
}

// ToContext adds the Store to this context if it supports
// the Setter interface.
func ToContext(c Setter, state *State) {
	c.Set(key, state)
}
