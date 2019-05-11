package param

import (
	"github.com/go-openapi/spec"
	"github.com/google/uuid"
	"github.com/mruck/athena/goFuzz/util"
)

// State for mutating a parameter
type State struct {
	// We need a way of mapping a parameter to its state
	// so embed the swagger parameter metadata
	spec.ParamProps
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
func New(param spec.ParamProps) *State {
	return &State{ParamProps: param}
}

// InitializeParamState takes a list of spec.Parameter objects and returns
// a list of *State objects for each parameter
func InitializeParamState(params []spec.Parameter) []*State {
	// Allocate our list
	state := make([]*State, len(params))
	for i, param := range params {
		state[i] = New(param.ParamProps)
	}
	return state
}

// TODO: keep track of what enums we've tried
// Add enumIndex or look in previous values?
func mutateEnum(schema spec.Schema) interface{} {
	randIndex := len(schema.Enum) % int(uuid.New().ID())
	return schema.Enum[randIndex]
}

func mutateBySchema(schema spec.Schema) interface{} {
	if schema.Enum != nil {
		return mutateEnum(schema)
	}
	dataType := schema.Type[0]
	if dataType == "object" {
		obj := map[string]interface{}{}
		for key, schema := range schema.Properties {
			obj[key] = mutateBySchema(schema)
		}
		return obj
	}
	//if dataType == "array" {
	//	obj := []interface{}{}
	//	obj[0] = mutateBySchema(schema.Items.Properties)
	//	return obj
	//}
	return util.Rand(schema.Type)
}

// Mutate sets State.Next based on the information provided in
// spec.Schema
// TODO: how to handle random values for array type for path params, etc
// Add a test for this
// TODO: store previous value
func (param *State) Mutate() {
	param.Next = mutateBySchema(*param.Schema)
}
