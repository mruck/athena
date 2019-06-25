package param

import (
	"github.com/go-openapi/spec"
	"github.com/mruck/athena/goFuzz/swagger"
)

// Param state for mutating a parameter
type Param struct {
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
func New(param spec.Parameter) *Param {
	return &Param{Parameter: param}
}

// InitializeParamState takes a list of spec.Parameter objects and returns
// a list of *State objects for each parameter
func InitializeParamState(params []spec.Parameter) []*Param {
	// Allocate our list
	state := make([]*Param, len(params))
	for i, param := range params {
		state[i] = New(param)
	}
	return state
}

// MockData generates dummy data for testing only
func (param *Param) MockData() {
	param.Next = swagger.MockAny(&param.Parameter)
}
