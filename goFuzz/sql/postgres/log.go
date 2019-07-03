package postgres

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mruck/athena/lib/log"
	"github.com/mruck/athena/lib/util"
	"github.com/pkg/errors"
)

// For indexing into postgres csv:
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

// jsonifiedQuery is for converting a query in array form to struct form
type jsonifiedQuery struct {
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

// LogPathEnvVar contains the postgres log path.
const LogPathEnvVar = "POSTGRES_LOG_PATH"

// LogPath is the default path to the postgres path, configurable at start up of pg container
// via `log_filename` parameter, but this is a bit weird cause its for the log file, but uses
// .csv for csv output
const LogPath = "/var/log/athena/postgres/postgres.csv"

// triaged postgres errors are written to this file
const triagedLogFile = "triaged_postgres.log"

// Postgres message severity levels taken from
// https://www.postgresql.org/docs/9.2/runtime-config-logging.html
// Table 18-1. Message Severity Levels
const (
	postgresError   = "ERROR"
	postgresPanic   = "PANIC"
	postgresFatal   = "Fatal"
	postgresWarning = "Warning"
)

// PGLog is responsible for reading the postgres log file
// at `path` starting from `lastTimeStamp`
type PGLog struct {
	// last query read had this timestamp
	lastTimeStamp string
	// path to postgres log file
	path string
	// triaged postgres log
	triagedLog *os.File
	// postgres log is a csv, each csv is loaded as []string
	queryMetadata [][]string
}

// NewLog takes in the path to the postgres log and returns a postgres
// load reader
func NewLog() *PGLog { // Open a file for logging triaged postgres errors
	name := filepath.Join(util.GetLogPath(), triagedLogFile)
	fp, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	util.Must(err == nil, "%+v\n", errors.WithStack(err))

	pgLog := &PGLog{
		path:       getPostgresLogPath(),
		triagedLog: fp,
	}
	return pgLog
}

// Seek over all stale data so that we are pointing to the most recent
// queries
func (pglog *PGLog) Seek() {
	_, err := pglog.Next()
	util.Must(err == nil, "%+v\n", errors.WithStack(err))
}

// Next reads the postgres queries starting at `timestamp`, extracts the raw queries
// from the meta data for each query, and returns them
func (pglog *PGLog) Next() ([]string, error) {
	// Reset stale data
	pglog.queryMetadata = [][]string{}

	// Read the postgres log
	records, err := util.LoadCSVFile(pglog.path)
	if err != nil {
		return nil, err
	}

	// Truncate so only read the most recent records
	truncated, err := truncate(pglog.lastTimeStamp, records)
	if err != nil {
		return nil, err
	}
	pglog.queryMetadata = truncated

	// We read something new, update latest timestamp
	if truncated != nil {
		pglog.lastTimeStamp = records[len(truncated)-1][LogTime]
	}

	// Extract raw queries
	raw := pglog.extractRawQueries()
	return raw, nil
}

func toStruct(query []string) jsonifiedQuery {
	return jsonifiedQuery{
		LogTime:       query[LogTime],
		ErrorSeverity: query[ErrorSeverity],
		SQLStateCode:  query[SQLStateCode],
		Message:       query[Message],
		Detail:        query[Detail],
		Hint:          query[Hint],
		InternalQuery: query[InternalQuery],
		Context:       query[Context],
		Query:         query[Query],
	}

}

// TODO: pg emits this error msg a lot. right now i'm just
// ignoring it but eventually I should figure it out and fix it
const vagrantMsg = "role \"vagrant\" does not exist"

// Triage the postgres log for hints, errors, etc
func (pglog *PGLog) Triage() {
	for _, query := range pglog.queryMetadata {
		isErr := isPostgresError(query[ErrorSeverity])
		// Nothing went wrong
		if !isErr {
			continue
		}
		if query[Message] == vagrantMsg {
			continue
		}
		data := toStruct(query)
		JSONData, err := json.Marshal(data)
		if err != nil {
			log.Errorf("Failed to triage postgres log: %+v", errors.WithStack(err))
			return
		}
		_, err = pglog.triagedLog.Write(append(JSONData, '\n'))
		if err != nil {
			log.Errorf("Failed to triage postgres log: %+v", errors.WithStack(err))
			return
		}
	}
}

// Sanitize the query emitted by postgres log.
// Postgres logs queries with leading characters and double quotes like:
// "statement: SELECT  \"users\".* FROM \"users\" WHERE \"users\".\"username_lower\" = 'd0f815' LIMIT 1"
// This causes the sql parser to return an error.  Sanitize so that it looks like:
// "SELECT  users.* FROM users WHERE users.username_lower = 'd0f815' LIMIT 1"
func sanitize(query string) string {
	// Trim leading prefix up to semicolon
	trimmed := strings.SplitN(query, ":", 2)
	if len(trimmed) == 2 {
		// There was something to trim and the query is the 2nd element
		query = trimmed[1]
	} else {
		// There was nothing to trim
		query = trimmed[0]
	}

	// Remove double quotes
	query = strings.Replace(query, "\"", "", -1)

	// Trim leading/trailing whitespace
	return strings.Trim(query, " ")
}

// extractRawQueries extracts the raw sql queries from the `message` field of
// each query metadata object.
// messages that are not queries) and queries that errored out.  All queries are
// prefixed with `statement`, so be sure to remove that, i.e.:
// "statement: create table cities (name varchar(80), temp int);"
func (pglog *PGLog) extractRawQueries() []string {
	rawQueries := []string{}

	// Extract the raw query
	for _, query := range pglog.queryMetadata {
		// Query errored out
		if isPostgresError(query[ErrorSeverity]) {
			continue
		}
		rawQueries = append(rawQueries, sanitize(query[Message]))
	}
	return rawQueries
}

// isPostgresError checks postgres message severity levels
// and returns whether or not anything errored out
func isPostgresError(err string) bool {
	if err == postgresError || err == postgresWarning {
		return true
	}
	if err == postgresPanic || err == postgresFatal {
		log.Fatal("Postgres logged panic/fatal error")
	}
	return false
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
