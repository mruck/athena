package sql

// CRUD action on db
type CRUD int

// CRUD Sql operations
const (
	Update CRUD = iota
	Select
	Delete
	Insert
)

// TaintedQuery is a sql query tainted with user controlled data
type TaintedQuery struct {
	Param string
	// Raw query to run to get comparable results
	Query  string
	Table  string
	Column string
	CRUD   CRUD
}
