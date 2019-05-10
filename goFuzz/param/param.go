package param

// State for mutating a parameter
type State struct {
	CurrentValue   interface{}
	PreviousValues []interface{}
	HarValues      []string
	// Query to run to retrieve this value
	Query string
	// Table the value maps to (in case the query fails, just pop something from here)
	Table  string
	Column string
}
