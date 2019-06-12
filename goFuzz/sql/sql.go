package sql

import (
	"fmt"
	"strings"

	"github.com/mruck/athena/lib/log"
	"github.com/mruck/athena/lib/util"
)

// CheckForSQLInj updates AnalyzedLog.VulnerableSQL
func CheckForSQLInj(queries []string, params map[string]string) {
	// Log to file
}

// Search for user tainted queries
func Search(queries []string, params map[string]string) []TaintedQuery {
	if len(queries) == 0 || len(params) == 0 {
		return nil
	}
	taintedQueries := []TaintedQuery{}
	for _, query := range queries {
		for name, val := range params {
			// Do a simple string check before searching
			if !strings.Contains(query, val) {
				continue
			}
			log.Infof("Matched param \"%s\" with value \"%s\" in query:\n%v", name, val, query)
			taintedQuery, err := parseQuery(query, val)
			if err != nil {
				err = fmt.Errorf("error searching for param %s with value %v in query:\n%s\n%+v", name, val, query, err)
				log.Error(err)
				continue
			}
			if taintedQuery != nil {
				taintedQuery.Name = name
				log.Infof("Tainted query:")
				util.PrettyPrintStructInfo(taintedQuery)
				taintedQueries = append(taintedQueries, *taintedQuery)
			}
		}
	}
	return taintedQueries
}
