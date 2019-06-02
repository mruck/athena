package sql

// Log contains a reader to sql log (in this case postgres), and provides
// methods on it for analysis
type Log struct {
	Reader         *PostgresLog
	taintedQueries []TaintedQuery
	// Vulnerable SQL
	// SQL errors
	// Hints
}

// NewLog allocates a new sql log object
func NewLog() *Log {
	return &Log{Reader: NewPostgresLog()}
}

// Analyze SQL log dumped by postgres (or in the future another db) by reading
// from the log then triaging the queries
func (log *Log) Analyze(params []string) error {
	// Reset stale data
	log.queryMetadata = [][]string{}
	log.taintedQueries = []TaintedQuery{}

	var err error
	// Read queries dumped by PG
	log.queryMetadata, err = log.Reader.Next()
	if err != nil {
		return err
	}

	// No new queries, nothing else to do here
	if len(log.queryMetadata) == 0 {
		return nil
	}

	// Extract queries from query metadata
	rawQueries := log.extractRawQueries()

	// Search for present params
	log.taintedQueries, err = search(params, rawQueries)
	if err != nil {
		return err
	}

	// We have tainted queries, check for sql injection
	if len(log.taintedQueries) > 0 {
		log.validateSQL()
	}

	// Check for hints, errors, etc from postgres
	log.triageErrors()

	return nil
}

// validateSQL updates AnalyzedLog.VulnerableSQL
func (log *Log) validateSQL() {
	// Log to file
}

// triageErrors updates SQL errors by parsing any errors, hints, etc from the log
func (log *Log) triageErrors() {
	for _, query := range log.queryMetadata {
		if isPostgresError(query[ErrorSeverity]) {
			continue
		}
	}
}

// GetTaintedQueries postgres log for user tainted queries
func (reader *PostgresLog) GetTaintedQueries(params []string) ([]TaintedQuery, error) {
	return nil, nil
}

// CheckForSQLInj checks if the most recent queries are vulnerable to sql inj
func (reader *PostgresLog) CheckForSQLInj(params []string) (bool, error) {
	return false, nil
}
