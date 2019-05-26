package sql

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

func New(pathToFile string) *PostgresLogReader {
	return nil
}

func (reader *PostgresLogReader) Next() ([]QueryMeta, error) {
	return nil, nil
}
