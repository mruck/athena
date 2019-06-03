package postgres

import (
	"fmt"
	"os"
	"strings"

	"github.com/mruck/athena/lib/util"
	"github.com/pkg/errors"
)

// For indexing into postgres csv, taken from:
// https://www.postgresql.org/docs/9.2/runtime-config-logging.html#RUNTIME-CONFIG-LOGGING-WHERE
const (
	LogTime       = 0
	ErrorSeverity = 11
	SQLStateCode  = 12
	Message       = 13
	Detail        = 14
	Hint          = 15
	InternalQuery = 16
	Context       = 18
	Query         = 19
)

// LogPathEnvVar contains the postgres log path.
const LogPathEnvVar = "POSTGRES_LOG_PATH"

// LogPath is the default path to the postgres path, configurable at start up of pg container
// via `log_filename` parameter, but this is a bit weird cause its for the log file, but uses
// .csv for csv output
const LogPath = "/var/log/athena/postgres/postgres.csv"

// Postgres message severity levels taken from
// https://www.postgresql.org/docs/9.2/runtime-config-logging.html
// Table 18-1. Message Severity Levels
const (
	postgresError   = "ERROR"
	postgresPanic   = "PANIC"
	postgresFatal   = "Fatal"
	postgresWarning = "Warning"
)

// Log is responsible for reading the postgres log file
// at `path` starting from `lastTimeStamp`
type Log struct {
	// last query read had this timestamp
	lastTimeStamp string
	// path to postgres log file
	path string
	// parsedErrors from postgres log are dumped to a file
	parsedErrors *os.File
	// postgres log is a csv, each csv is loaded as []string
	queryMetadata [][]string
}

// NewLog takes in the path to the postgres log and returns a postgres
// load reader
func NewLog() *Log {
	return &Log{path: getPostgresLogPath()}
}

// Next reads the postgres queries starting at `timestamp`, extracts the raw queries
// from the meta data for each query, and returns them
func (log *Log) Next() ([]string, error) {
	// Reset stale data
	log.queryMetadata = [][]string{}

	// Read the postgres log
	records, err := util.LoadCSVFile(log.path)
	if err != nil {
		return nil, err
	}

	// Truncate so only read the most recent records
	truncated, err := truncate(log.lastTimeStamp, records)
	if err != nil {
		return nil, err
	}
	log.queryMetadata = truncated

	// We read something new, update latest timestamp
	if truncated != nil {
		last := len(truncated) - 1
		log.lastTimeStamp = records[last][LogTime]
	}

	// Extract raw queries
	raw := log.extractRawQueries()
	return raw, nil
}

// Triage the postgres log for hints, errors, etc
func (log *Log) Triage() error {
	// Log hints, errors, etc to file
	return nil
}

// extractRawQueries extracts the raw sql queries from the `message` field of
// each query metadata object.  Skip information messages (i.e. postgres start up
// messages that are not queries) and queries that errored out.  All queries are
// prefixed with `statement`, so be sure to remove that, i.e.:
// "statement: create table cities (name varchar(80), temp int);"
func (log *Log) extractRawQueries() []string {
	rawQueries := []string{}

	// Extract the raw query
	for _, query := range log.queryMetadata {

		// Query errored out
		if isPostgresError(query[ErrorSeverity]) {
			continue
		}

		// This is an informational message, not a query
		if !strings.HasPrefix(query[Message], "statement:") {
			continue
		}

		// Remove prefix
		trimmed := strings.TrimPrefix(query[Message], "statement: ")
		rawQueries = append(rawQueries, trimmed)
	}
	return rawQueries
}

// isPostgresError checks postgres message severity levels
// and returns whether or not anything errored out
func isPostgresError(err string) bool {
	return err == postgresError ||
		err == postgresPanic ||
		err == postgresFatal ||
		err == postgresWarning
}

// Truncate returns the list of records appended since the
// most recent time stamp
func truncate(timestamp string, records [][]string) ([][]string, error) {
	// This is the first time we've read the log file
	if timestamp == "" {
		return records, nil
	}
	for i, record := range records {
		if record[LogTime] == timestamp {
			return records[i+1:], nil
		}
	}
	return nil, errors.WithStack(fmt.Errorf("unable to find timestamp %v in query list", timestamp))
}

// getPostgresLogPath returns the path to the postgres log set in the env, or defaults
// to PostgresLogPath
func getPostgresLogPath() string {
	path := os.Getenv(LogPathEnvVar)
	if path == "" {
		return LogPath
	}
	return path
}
