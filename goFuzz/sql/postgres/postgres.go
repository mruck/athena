package postgres

import (
	"fmt"

	"github.com/mruck/athena/lib/util"
)

// Postgres instance with logs and an active connection for querying
type Postgres struct {
	Log  *PGLog
	Conn *Connection
}

// New returns a new postgres object
func New() *Postgres {
	return &Postgres{
		Log:  NewLog(),
		Conn: NewConnection(getConnStr()),
	}
}

// Get postgres connection string
func getConnStr() string {
	user := util.DefaultEnv("TARGET_DB_USER", "root")
	password := util.DefaultEnv("TARGET_DB_PASSWORD", "mysecretpassword")
	dbname := util.DefaultEnv("TARGET_DB_NAME", "fuzz_db")
	port := util.DefaultEnv("TARGET_DB_PORT", "5432")
	host := util.DefaultEnv("TARGET_DB_HOST", "localhost")
	connStr := "dbname=%s user=%s password=%s port=%s host=%s sslmode=disable"
	return fmt.Sprintf(connStr, dbname, user, password, port, host)
}
