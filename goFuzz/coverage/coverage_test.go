package coverage

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/mruck/athena/lib/util"
	"github.com/stretchr/testify/require"
)

func TestReadCoverage1(t *testing.T) {
	readCoverageInner(t, "coverage1.json", "coverage2.json")
}

func TestReadCoverage2(t *testing.T) {
	readCoverageInner(t, "dummy1.json", "dummy2.json")
}

func readCoverageInner(t *testing.T, file1 string, file2 string) {
	// Tmpfile for writing coverage to
	tmp, err := ioutil.TempFile("/tmp", "cov-")
	require.NoError(t, err)
	//defer os.Remove(tmp.Name())

	coverage := New(tmp.Name())
	err = util.CopyFile(tmp.Name(), file1)
	require.NoError(t, err)

	// Read the coverage
	err = coverage.Update()
	require.NoError(t, err)

	// Check what we got
	require.True(t, coverage.Delta > 0)
	require.True(t, coverage.Cumulative > 0)
	oldDelta := coverage.Delta
	oldCumulative := coverage.Cumulative
	fmt.Printf("Delta percentage: %v\n", coverage.Delta)
	fmt.Printf("Cumulative percentage: %v\n", coverage.Cumulative)

	// Update coverage
	err = util.CopyFile(tmp.Name(), file2)
	require.NoError(t, err)

	// Read coverage again
	err = coverage.Update()
	require.NoError(t, err)

	fmt.Printf("Delta percentage: %v\n", coverage.Delta)
	fmt.Printf("Cumulative percentage: %v\n", coverage.Cumulative)

	// Check we got more coverage
	require.True(t, coverage.Delta > oldDelta)
	require.True(t, coverage.Cumulative > oldCumulative)
}
