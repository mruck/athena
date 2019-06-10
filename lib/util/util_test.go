package util

import (
	"testing"

	"github.com/mruck/athena/lib/log"
	"github.com/stretchr/testify/require"
)

//func TestReadLines(t *testing.T) {
//   fp, err := os.Open("test/src_line_coverage")
//   require.NoError(t, err)
//   lines, err := ReadLines(fp)
//   require.NoError(t, err)
//   for _, line := range lines {
//   	log.Infof(line)
//   }

func TestStringify(t *testing.T) {
	float := 5.123
	require.Equal(t, "5.123", Stringify(float))
	letters := []string{"a", "b", "c"}
	log.Infof(Stringify(letters))
}

// Test unmarshalling an empty file
func TestUnmarshalFileEmpty(t *testing.T) {
	type Test struct {
		A string
	}
	test := &Test{}
	err := UnmarshalFile("foo.json", test)
	log.Info(err)
	log.Info(test)
}

func TestReadLineByLine(t *testing.T) {
	correctlines := []string{"line1", "line2"}
	lines, err := ReadFileLineByLine("test/lines.txt")
	require.NoError(t, err)
	for i, line := range lines {
		require.Equal(t, correctlines[i], line)
	}
}
