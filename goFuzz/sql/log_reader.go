package sql

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

// PostgresLogEnvVar contains the postgres log path.
const PostgresLogEnvVar = "POSTGRES_LOG_PATH"

// PostgresLogPath is the default path to the postgres path, configurable at start up of pg container
// via `log_filename` parameter, but this is a bit weird cause its for the log file, but uses
// .csv for csv output
const PostgresLogPath = "/var/log/athena/postgres.csv"

// Postgres message severity levels taken from
// https://www.postgresql.org/docs/9.2/runtime-config-logging.html
// Table 18-1. Message Severity Levels
const (
	PostgresErr     = "ERROR"
	PostgresPanic   = "PANIC"
	PostgresFatal   = "Fatal"
	PostgresWarning = "Warning"
)

// PostgresLog is responsible for reading the postgres log file
// at `path` starting from `lastTimeStamp`
type PostgresLog struct {
	lastTimeStamp string
	path          string
}

// NewPostgresLog takes in the path to the postgres log and returns a postgres
// load reader
func NewPostgresLog() *PostgresLog {
	return &PostgresLog{path: getPostgresLogPath()}
}

// Next returns the most recent queries run by postgres
func (reader *PostgresLog) Next() ([][]string, error) {
	// Read the postgres log
	records, err := util.LoadCSVFile(reader.path)
	if err != nil {
		return nil, err
	}

	// Truncate so only read the most recent records
	truncated, err := truncate(reader.lastTimeStamp, records)
	if err != nil {
		return nil, err
	}

	// We read something new, update latest timestamp
	if truncated != nil {
		last := len(truncated) - 1
		reader.lastTimeStamp = records[last][LogTime]
	}

	return truncated, nil
}

// isPostgresError checks postgres message severity levels
// and returns whether or not anything errored out
func isPostgresError(err string) bool {
	return err == PostgresErr ||
		err == PostgresPanic ||
		err == PostgresFatal ||
		err == PostgresWarning
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
	path := os.Getenv(PostgresLogEnvVar)
	if path == "" {
		return PostgresLogPath
	}
	return path
}
