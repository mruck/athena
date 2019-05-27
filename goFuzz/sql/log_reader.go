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

//func toQueryMeta(records [][]string) []QueryMeta {
//	metas := make([]QueryMeta, len(records))
//	for i, record := range records {
//		metas[i].LogTime = record[LogTime]
//		metas[i].ErrorSeverity = record[ErrorSeverity]
//		metas[i].ErrorSeverity = record[ErrorSeverity]
//	}
//	return metas
//}

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

// PostgresLogReader is responsible for reading the postgres log file
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

// Next returns the most recent queries run by postgres
func (reader *PostgresLogReader) Next() ([][]string, error) {
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