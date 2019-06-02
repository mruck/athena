package sql

// Log contains a reader to sql log (in this case postgres), and provides
// methods on it for analysis
type Log struct {
	Reader         *PostgresLogReader
	queryMetadata  [][]string
	taintedQueries []TaintedQuery
	// Vulnerable SQL
	// SQL errors
	// Hints
}

// NewLog allocates a new sql log object
func NewLog() *Log {
	return &Log{Reader: NewPostgresLogReader()}
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
	// Log PG errs to file
}

// extractRawQueries extracts the raw sql queries from each query metadata
// object, skipping queries that errored out
func (log *Log) extractRawQueries() []string {
	rawQueries := []string{}
	// Extract the raw query
	for _, query := range log.queryMetadata {
		if isPostgresError(query[ErrorSeverity]) {
			continue
		}
		rawQueries = append(rawQueries, query[Message])
	}
	return rawQueries
}
