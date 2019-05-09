package util

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadLines(t *testing.T) {
	fp, err := os.Open("test/src_line_coverage")
	require.NoError(t, err)
	lines, err := ReadLines(fp)
	require.NoError(t, err)
	for _, line := range lines {
		fmt.Println(line)
	}
}
