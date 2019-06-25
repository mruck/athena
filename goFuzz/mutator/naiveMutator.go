package mutator

import (
	"github.com/mruck/athena/goFuzz/param"
	"github.com/mruck/athena/goFuzz/route"
	"github.com/mruck/athena/goFuzz/swagger"
)

// MutateRoute mutates the parameters on a given route.
// Setting param.Next for each parameter, or nil if the paramater shouldn't
// be sent
func (mutator *Mutator) MutateRoute(route *route.Route) {
	for _, param := range route.Params {
		mutateParam(param)
	}
}

// Mutate a body parameter
func mutateBody(param *param.Param) {
}

// Mutate a primitive parameter (path, query)
func mutatePrimitive(param *param.Param) {
	//var val interface{}
	//if param.Type == "array" {
	//	val = generatePrimitiveArray(param.Items)
	//}
	//if param.Enum != nil {
	//	val = generateEnum(param.Enum)
	//}
	//return util.Rand(param.Type)
}

func mutateParam(param *param.Param) {
	// Mutate the leaves
	if param.In == "body" {
		mutateBody(param)
	} else {
		mutatePrimitive(param)
	}
	// Correctly format the data (i.e. into json)
	param.Next = swagger.MockAny(&param.Parameter)

}
