package postgres

import (
	"fmt"
	"os"

	"github.com/mruck/athena/lib/util"
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
const LogPath = "/var/log/athena/postgres.csv"

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
	// ParsedErrors from postgres log are dumped to a file
	ParsedErrors *os.File
	// postgres log is a csv, each csv is loaded as []string
	queryMetadata [][]string
}

// NewLog takes in the path to the postgres log and returns a postgres
// load reader
func NewLog() *Log {
	return &Log{path: getPostgresLogPath()}
}

// Next returns the most recent queries run by postgres
func (log *Log) Next() ([][]string, error) {
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

	// We read something new, update latest timestamp
	if truncated != nil {
		last := len(truncated) - 1
		log.lastTimeStamp = records[last][LogTime]
	}

	return truncated, nil
}

// ExtractRawQueries extracts the raw sql queries from each query metadata
// object, skipping queries that errored out
func (log *Log) ExtractRawQueries() []string {
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

// Triage the postgres log for hints, errors, etc
func (log *Log) Triage() error {
	// Log hints, errors, etc
	return nil
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
	return nil, fmt.Errorf("unable to find timestamp %v in query list", timestamp)
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
