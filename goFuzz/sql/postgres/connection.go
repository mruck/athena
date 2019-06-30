package postgres

import (
	"database/sql"
	"log"

	"github.com/pkg/errors"
)

// Connection to a postgres database
type Connection struct {
	db *sql.DB
}

// NewConnection connects to a postgres database.  Panic on error.
func NewConnection(connStr string) *Connection {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(errors.WithStack(err))
	}
	return &Connection{db: db}
}

// LookUp the table and column and return a valid value, or nil if not found
func (conn *Connection) LookUp(table string, col string) interface{} {
	return nil
}
