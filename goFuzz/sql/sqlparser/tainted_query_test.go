package sqlparser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatchParam(t *testing.T) {
	query := "SELECT  users.* FROM users WHERE users.username_lower = 'system' LIMIT 1"
	param := "system"
	matched := matchParam(param, query)
	require.True(t, matched)

	param = "users"
	matched = matchParam(param, query)
	require.False(t, matched)
}
