package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRand(t *testing.T) {
	val := Rand("integer")
	_, ok := val.(int)
	require.True(t, ok)

	val = Rand("number")
	_, ok = val.(int)
	require.True(t, ok)

	val = Rand("boolean")
	_, ok = val.(bool)
	require.True(t, ok)

	val = Rand("decimal")
	_, ok = val.(float32)
	require.True(t, ok)

	val = Rand("string")
	_, ok = val.(string)
	require.True(t, ok)
}
