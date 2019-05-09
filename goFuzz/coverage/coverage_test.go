package coverage

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func copyFile(dst string, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()

}

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
	err = copyFile(tmp.Name(), file1)
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
	err = copyFile(tmp.Name(), file2)
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
