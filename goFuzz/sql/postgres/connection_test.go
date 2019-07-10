package postgres

import (
	"testing"

	"github.com/mruck/athena/lib/log"
	"github.com/stretchr/testify/require"
)

const postgresRunning = "For local testing, ensure postgres is running at:\n" +
	"docker run -d --rm -p 5432:5432 -e POSTGRES_USER=root postgres:10.5"

func TestConnection(t *testing.T) {
	// Sanity check PG is running
	log.Info(postgresRunning)

	// Test that we can connect to PG
	connStr := getConnStr()
	conn := NewConnection(connStr)
	require.NotNil(t, conn)

	// Create a table
	tablename, err := conn.mockTable()
	require.NoError(t, err)

	// Write fake data
	err = conn.mockInsert(tablename)
	require.NoError(t, err)

	// Test reading data
	val := conn.LookUp(tablename, "name")
	require.NoError(t, err)
	require.Equal(t, "sunnyvale", val)

	// Test trying to read data that doesn't exit
	val = conn.LookUp("noSuchTable", "name")
	require.Nil(t, val)
}
