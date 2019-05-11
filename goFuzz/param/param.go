package param

import "github.com/go-openapi/spec"

// State for mutating a parameter
type State struct {
	// We need a way of mapping a parameter to its state
	Param          spec.Parameter
	CurrentValue   interface{}
	PreviousValues []interface{}
	HarValues      *[]string
	// Query to run to retrieve this value
	Query string
	// Table the value maps to (in case the query fails, just pop something from here)
	Table  string
	Column string
}

// New allocates a new parameter state object
func New(param spec.Parameter) *State {
	return &State{Param: param}
}

// InitializeParamState takes a list of spec.Parameter objects and returns
// a list of *State objects for each parameter
func InitializeParamState(params []spec.Parameter) []*State {
	// Allocate our list
	state := make([]*State, len(params))
	for i, param := range params {
		state[i] = New(param)
	}
	return state
}
