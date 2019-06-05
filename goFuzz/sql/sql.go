package sql

import (
	"fmt"
	"strings"
)

// CheckForSQLInj updates AnalyzedLog.VulnerableSQL
func CheckForSQLInj(queries []string, params []string) {
	// Log to file
}

// Search for user tainted queries
func Search(queries []string, params []string) ([]TaintedQuery, error) {
	if len(queries) == 0 || len(params) == 0 {
		return nil, nil
	}
	taintedQueries := []TaintedQuery{}
	for _, query := range queries {
		for _, param := range params {
			// Do a simple string check before searching
			if !strings.Contains(query, param) {
				continue
			}
			taintedQuery, err := parseQuery(query, param)
			if err != nil {
				err = fmt.Errorf("error parsing query: %s\n%+v", query, err)
				return nil, err
			}
			if taintedQuery != nil {
				taintedQueries = append(taintedQueries, *taintedQuery)
			}
		}
	}
	return taintedQueries, nil
}
