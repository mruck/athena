package sql

import (
	"fmt"
	"strings"
)

// CheckForSQLInj updates AnalyzedLog.VulnerableSQL
func CheckForSQLInj(queries []string, params map[string]string) {
	// Log to file
}

// Search for user tainted queries
func Search(queries []string, params map[string]string) ([]TaintedQuery, error) {
	if len(queries) == 0 || len(params) == 0 {
		return nil, nil
	}
	taintedQueries := []TaintedQuery{}
	for _, query := range queries {
		for name, val := range params {
			// Do a simple string check before searching
			if !strings.Contains(query, val) {
				continue
			}
			taintedQuery, err := parseQuery(query, val)
			if err != nil {
				err = fmt.Errorf("error parsing query: %s\n%+v", query, err)
				return nil, err
			}
			if taintedQuery != nil {
				taintedQuery.Name = name
				taintedQueries = append(taintedQueries, *taintedQuery)
			}
		}
	}
	return taintedQueries, nil
}
