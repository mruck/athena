package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMongoClient(t *testing.T) {
	cli, err := NewClient("localhost", "27017", "athena")
	require.NoError(t, err)
	require.NotNil(t, cli)

	val, err := cli.LookUp("exceptions")
	require.NoError(t, err)
	require.NotEqual(t, "", val)
}
