package coverage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadCoverage(t *testing.T) {
	coverage := New("coverage.json")
	err := coverage.Update()
	require.NoError(t, err)
	require.True(t, coverage.Delta > 0)
	require.True(t, coverage.Cumulative > 0)
	fmt.Printf("Delta percentage: %v\n", coverage.Delta)
	fmt.Printf("Cumulative percentage: %v\n", coverage.Cumulative)
}
