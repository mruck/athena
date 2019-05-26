package sql

type Parser struct {
	// Path to python script for parsing sql
	scriptPath string
}

// AnalyzeSQL searcjes for parameters in the given queries
func (parser *Parser) AnalyzeSQL(params []string, queries []string) ([]ParamQuery, error) {
	return nil, nil
}

// ParamQuery data matching a parameter to a query
type ParamQuery struct {
	Param string
	// Raw query to run to get comparable results
	Query  string
	Table  string
	Column string
	// CRUD action
	Method string
}
