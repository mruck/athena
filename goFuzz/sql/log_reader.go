package sql

import (
	"fmt"

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

// QueryMeta contains meta data about each query logged by postgres
type QueryMeta struct {
	LogTime       string
	ErrorSeverity string
	SQLStateCode  string
	Message       string
	Detail        string
	Hint          string
	InternalQuery string
	Context       string
	Query         string
}

type QueryMetas []QueryMeta

// PostgresLogRead is responsible for reading the postgres log file
// at `path` starting from `lastTimeStamp`
type PostgresLogReader struct {
	lastTimeStamp string
	path          string
}

// New takes in the path to the postgres load and returns a postgres
// load reader
func New(path string) *PostgresLogReader {
	return &PostgresLogReader{path: path}
}

// Truncate returns the list of records appended since the
// most recent time stamp
func truncate(timestamp string, records [][]string) [][]string {
	// This is the first time we've read the log file
	if timestamp == "" {
		return records
	}
	for i, record := range records {
		if record[LogTime] == timestamp {
			return records[i:]
		}
	}
	return nil
}

func (reader *PostgresLogReader) Next() ([]QueryMeta, error) {
	// Read the postgres log
	records := util.LoadCSVFile(reader.path)

	// Truncate so only read the most recent records
	truncated := truncate(reader.lastTimeStamp, records)

	if truncated == nil {
		return nil, fmt.Errorf("unable to find timestamp in query list")
	}
	// Convert each record to query meta object

	return nil, nil
}
