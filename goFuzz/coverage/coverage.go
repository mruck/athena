package coverage

import (
	"os"

	"github.com/mruck/athena/goFuzz/util"
)

// Coverage contains metadata abouta  coverage object
type Coverage struct {
	Cumulative          float64
	Delta               float64
	CoverageFilePointer *os.File
}

// New returns a coverage object that reads from coveragePath
func New(coveragePath string) *Coverage {
	fp, err := os.Open(coveragePath)
	util.PanicIfErr(err)
	return &Coverage{Cumulative: 0, Delta: 0, CoverageFilePointer: fp}
}

// ReadCoverage reads new coverage from the associated fp and updates the
// cumulative and delta coverage values
func (coverage *Coverage) ReadCoverage() {

}
