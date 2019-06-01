package sql

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
	// Create a tmp filel for the csv reader
	tmp, err := ioutil.TempFile("/tmp", "")
	require.NoError(t, err)
	defer os.Remove(tmp.Name())

	// Copy csv contents into it
	err = util.CopyFile(tmp.Name(), cities1)
	require.NoError(t, err)

	pgReader := NewPostgresLogReader(tmp.Name())

	// Read in all records with no time stamp
	records, err := pgReader.Next()
	require.NoError(t, err)

	for _, record := range records {
		// These fields should always be present
		require.NotNil(t, record[LogTime])
		require.NotNil(t, record[Message])
	}

	// Check that the last ts is present
	last := len(records) - 1
	ts := "2019-05-27 14:47:57.840 UTC"
	require.Equal(t, ts, records[last][LogTime])

	// Update the csv and read again
	err = util.CopyFile(tmp.Name(), cities2)
	require.NoError(t, err)

	// Read in all records, this time with a time stamp
	records, err = pgReader.Next()
	require.NoError(t, err)

	for _, record := range records {
		// Check that other records are present
		require.NotNil(t, record[LogTime])
		require.NotNil(t, record[Message])
		// Check that the old time stamp isn't present (we should start
		// reading after it)
		require.NotEqual(t, ts, record[LogTime])
	}
}
