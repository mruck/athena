package sqlparser

// Action on database
type Action int

// CRUD Sql operations
const (
	Update Action = iota
	Select
	Delete
	Insert
	Create
	Truncate
	Drop
	Alter
	Rename
)

// TaintedQuery is a sql query tainted with user controlled data
type TaintedQuery struct {
	// Parameter value that we searched for and identified inside the query
	Param string
	// Raw query to run to get comparable results
	Query  string
	Table  string
	Column string
	Action Action
}
