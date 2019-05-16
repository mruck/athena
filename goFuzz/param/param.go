package param

import (
	"github.com/go-openapi/spec"
	"github.com/mruck/athena/goFuzz/swagger"
)

// State for mutating a parameter
type State struct {
	// We need a way of mapping a parameter to its state
	// so embed the swagger parameter metadata
	spec.Parameter
	// Next value to send
	Next           interface{}
	PreviousValues *[]interface{}
	HarValues      *[]string
	// Query to run to retrieve this value
	Query string
	// Table the value maps to (in case the query fails, just pop something from here)
	Table  string
	Column string
}

// New allocates a new parameter state object
func New(param spec.Parameter) *State {
	return &State{Parameter: param}
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

// Mutate sets State.Next based on the information provided in
// spec.Schema
// TODO: how to handle random values for array type for path params, etc
// Add a test for this
// TODO: store previous value
func (param *State) Mutate() {
	param.Next = swagger.GenerateAny(&param.Parameter)
}
