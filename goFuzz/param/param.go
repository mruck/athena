package param

import (
	"fmt"

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

// GetMetadata returns a list of metadata objects embedded in the
// top leevl parameter
func (param *Param) GetMetadata() []*swagger.Metadata {
	return swagger.ReadAllMetadata(&param.Parameter)
}

// LatestValues returns the latest values sent for this param.  We say values
// because in a body param, we treat each leaf node as a separate param.
// Each value is stringified, not sure if we should do this or not.
func (param *Param) LatestValues() []string {
	latest := []string{}
	metadata := swagger.ReadAllMetadata(&param.Parameter)

	for _, data := range metadata {
		// Get the value
		val := data.Values[0]
		// stringify
		stringified := fmt.Sprintf("%v", val)
		latest = append(latest, stringified)
	}

	return latest
}
