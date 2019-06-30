package sqlparser

import (
	"fmt"
	"os"
	"strings"

	"github.com/mruck/athena/lib/log"
	"github.com/uber/makisu/lib/utils"
)

// CheckForSQLInj updates AnalyzedLog.VulnerableSQL
func CheckForSQLInj(queries []string, params []string) {
	// Log to file
}

// whitelistErrors contains acceptable sql parsing errors
var whitelistErrors = []string{"COPY", "CREATE TABLE", "COMMENT ON COLUMN"}

// Parser contains global state about the sql parser
type Parser struct {
	TaintedQueries []*TaintedQuery
	// Total queries we attempted to parse
	TotalQueries int
	// Total number of queries sqlparser library failed on
	LibError int
	// Total number of queries Athena failed to handle
	AthenaError int
	// Log of queries Athena failed to handle
	AthenaErrorLog *os.File
	// Log of queries sqlparser errored on
	ParsingErrorLog *os.File
}

// NewParser returns a new parsing instance
func NewParser() *Parser {
	return &Parser{}
}

// PrettyPrint parser stats
func (parser *Parser) PrettyPrint() {
	log.Infof("Total queries attempted to parse: %d", parser.TotalQueries)
	log.Infof("Queries sqlparser library failed: %d", parser.LibError)
	log.Infof("Queries athena failed: %d", parser.AthenaError)
	log.Infof("Tainted queries: %d", len(parser.TaintedQueries))
}

// Search for user tainted queries
func (parser *Parser) Search(queries []string, params []string) ([]TaintedQuery, error) {
	errs := utils.NewMultiErrors()
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
			//log.Infof("Matched param \"%s\" with value \"%s\" in query:\n%v", name, val, query)
			taintedQuery, err := parseQuery(query, param)
			parser.TotalQueries++
			if err != nil {
				_ = parser.triageError(err, query, param)
				// We can't parse this query so don't bother
				break
			}
			if taintedQuery != nil {
				// Append for logging puroses
				parser.TaintedQueries = append(parser.TaintedQueries, taintedQuery)
				taintedQueries = append(taintedQueries, *taintedQuery)
			}
		}
	}
	return taintedQueries, errs.Collect()
}

// triageError checks if the error is in our whitelist of acceptable errors,
// emitting a warning if it's not severe, otherwise returning the err so it can
// be bubbled up
func (parser *Parser) triageError(err error, query string, param string) error {
	// This is an error in the library we are using
	if strings.Contains(err.Error(), LibErr) {
		parser.LibError++
	} else {
		parser.AthenaError++
	}
	err = fmt.Errorf("error searching for param value %v in query:\n%s\n%+v", param, query, err)
	return err
}
