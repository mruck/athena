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

// Get postgres connection string
func getConnStr() string {
	user := util.DefaultEnv("TARGET_DB_USER", "root")
	password := util.DefaultEnv("TARGET_DB_PASSWORD", "mysecretpassword")
	return fmt.Sprintf("user=%s password=%s", user, password)
}

// New returns a new postgres object
func New() *Postgres {
	return &Postgres{
		Log:  NewLog(),
		Conn: NewConnection(getConnStr()),
	}
}
