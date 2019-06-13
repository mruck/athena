package sql

import (
	"fmt"
	"strings"

	"github.com/mruck/athena/lib/log"
	"github.com/mruck/athena/lib/util"
	"github.com/uber/makisu/lib/utils"
)

// CheckForSQLInj updates AnalyzedLog.VulnerableSQL
func CheckForSQLInj(queries []string, params map[string]string) {
	// Log to file
}

// whitelistErrors contains acceptable sql parsing errors
var whitelistErrors = []string{"COPY"}

//var whitelistErrors = []string{"COPY", "CREATE TABLE"}

// triageError checks if the error is in our whitelist of acceptable errors,
// emitting a warning if it's not severe, otherwise returning the err so it can
// be bubbled up
func triageError(err error) error {
	// This is whitelisted, only emit warning
	if util.StringInSlice(err.Error(), whitelistErrors) {
		log.Warn(err)
		return nil
	}
	return err
}

// Search for user tainted queries
func Search(queries []string, params map[string]string) ([]TaintedQuery, error) {
	errs := utils.NewMultiErrors()
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
			log.Infof("Matched param \"%s\" with value \"%s\" in query:\n%v", name, val, query)
			taintedQuery, err := parseQuery(query, val)
			if err != nil {
				err = fmt.Errorf("error searching for param %s with value %v in query:\n%s\n%+v", name, val, query, err)
				err = triageError(err)
				if err != nil {
					errs.Add(err)
				}
				// We can't parse this query so don't bother
				break
			}
			if taintedQuery != nil {
				taintedQuery.Name = name
				log.Infof("Tainted query:")
				util.PrettyPrintStructInfo(taintedQuery)
				taintedQueries = append(taintedQueries, *taintedQuery)
			}
		}
	}
	return taintedQueries, errs.Collect()
}
