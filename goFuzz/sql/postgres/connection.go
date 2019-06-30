package postgres

import (
	"database/sql"
	"fmt"

	// Go Postgres driver for the database/sql package
	_ "github.com/lib/pq"

	"github.com/mruck/athena/lib/log"
	"github.com/mruck/athena/lib/util"
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

// LookUp returns a value at table and col or nil on error
func (conn *Connection) LookUp(table string, col string) interface{} {
	stmt := fmt.Sprintf("SELECT %s FROM %s LIMIT 1", col, table)
	row := conn.db.QueryRow(stmt)
	var val interface{}
	err := row.Scan(&val)
	if err != nil {
		// TODO: log err
		return nil
	}
	return val
}

// Generate table for testing. Returns the table name and any error
func (conn *Connection) mockTable() (string, error) {
	tableName := "table_" + util.RandString()[:4]
	stmt := fmt.Sprintf("CREATE TABLE %s (name varchar(255), temp int)", tableName)
	_, err := conn.db.Query(stmt)
	return tableName, err
}

func (conn *Connection) mockInsert(table string) error {
	// Build query string
	stmt := fmt.Sprintf("INSERT INTO %s VALUES ('sunnyvale', 80)", table)
	_, err := conn.db.Query(stmt)
	return err
}
