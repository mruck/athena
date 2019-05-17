package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

//func TestReadLines(t *testing.T) {
//   fp, err := os.Open("test/src_line_coverage")
//   require.NoError(t, err)
//   lines, err := ReadLines(fp)
//   require.NoError(t, err)
//   for _, line := range lines {
//   	fmt.Println(line)
//   }

func TestStringify(t *testing.T) {
	float := 5.123
	require.Equal(t, "5.123", Stringify(float))
	letters := []string{"a", "b", "c"}
	fmt.Println(Stringify(letters))
}
