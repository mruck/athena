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

// ReadMetadata reads the metadata objects embedded in the parameter
func (param *Param) ReadMetadata() []*swagger.Metadata {
	return swagger.ReadAllMetadata(&param.Parameter)
}

// StoreValue stores a single in the metadata object embedded
// in spec.Parameter
func (param *Param) StoreValue(val interface{}) {
	// Extract the embedded metadata struct
	metadata := swagger.ReadOneMetadata(&param.Parameter)
	// Insert the new value into index 0.  This is a pointer so the update
	// is done in place.
	metadata.Values = append([]interface{}{val}, metadata.Values...)
}
