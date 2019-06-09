package postgres

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mruck/athena/lib/util"
	"github.com/stretchr/testify/require"
)

const cities1 = "test/cities1.csv"
const cities2 = "test/cities2.csv"

func TestNext(t *testing.T) {
	// Create a tmp file for the csv reader
	tmp, err := ioutil.TempFile("/tmp", "")
	require.NoError(t, err)
	defer os.Remove(tmp.Name())

	// Copy csv contents into it
	err = util.CopyFile(tmp.Name(), cities1)
	require.NoError(t, err)

	// Set postgres log path in env
	os.Setenv(LogPathEnvVar, tmp.Name())
	pgReader := NewLog()

	// Read in all records with no time stamp
	rawQueries, err := pgReader.Next()
	require.NoError(t, err)

	// Check the raw queries we extracted.
	// The queries should match
	correctQueries := []string{
		"create table cities (name varchar(80), temp int);",
		"insert into cities (name, temp) values (sunnyvale, 60)\n;",
		"insert into cities (name, temp) values ('sunnyvale', 60);",
		"insert into cities (name, temp) values ('menlo park', 58);",
	}

	for i, rawQuery := range rawQueries {
		require.Equal(t, correctQueries[i], rawQuery)
	}

	// Check the query meta data
	records := pgReader.queryMetadata

	for _, record := range records {
		// These fields should always be present
		require.NotNil(t, record[LogTime])
		require.NotNil(t, record[Message])
		//	util.PrettyPrintStruct(record)
	}

	// Check that the last ts is present
	last := len(records) - 1
	ts := "2019-05-27 14:47:57.840 UTC"
	require.Equal(t, ts, records[last][LogTime])

	// Triage
	_ = pgReader.Triage()

	// Update the csv and read again
	err = util.CopyFile(tmp.Name(), cities2)
	require.NoError(t, err)

	// Read in all records, this time with a time stamp
	_, err = pgReader.Next()
	require.NoError(t, err)
	records = pgReader.queryMetadata

	for _, record := range records {
		// Check that other records are present
		require.NotNil(t, record[LogTime])
		require.NotNil(t, record[Message])
		// Check that the old time stamp isn't present (we should start
		// reading after it)
		require.NotEqual(t, ts, record[LogTime])
	}
	// Triage
	_ = pgReader.Triage()
}
