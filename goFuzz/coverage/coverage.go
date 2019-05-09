package coverage

import (
	"github.com/mruck/athena/goFuzz/util"
)

// Coverage contains metadata abouta  coverage object
type Coverage struct {
	// Total amount of coverage across all requests as a percentage
	Cumulative float64
	// New coverage received from most recent request as a percentage
	Delta float64
	// New coverage received from most recent request as a map
	//DeltaMap map[string][]*int
	FilePath string
	// Cumulative mapping of filepath to number of times lines
	// are hit.  Update after every request.
	Map map[string][]int
}

// New returns a coverage object that reads from coveragePath
func New(coveragePath string) *Coverage {
	return &Coverage{Cumulative: 0, Delta: 0, FilePath: coveragePath, Map: make(map[string][]int)}
}

// updateMap updates the cumulative coverage map with the most recent request's
// coverage and keeps track of the delta coverage in a map
func (coverage *Coverage) updateMap(newCoverage map[string][]int) map[string][]int {
	// Keep track of delta relative to cumulative coverage map
	deltaMap := make(map[string][]int)
	for newFilename, newLineCount := range newCoverage {
		// The filename is present. Update with new coverage
		if oldLineCount, ok := coverage.Map[newFilename]; ok {
			// Allocate an array for keeping track of delta line counts
			deltaLineCount := make([]int, len(oldLineCount))
			for i := range newLineCount {
				// This is unreachable
				if newLineCount[i] < 0 {
					deltaLineCount[i] = -1
					continue
				}
				// This is new coverage, add to the delta
				if oldLineCount[i] == 0 && newLineCount[i] > 0 {
					deltaLineCount[i] = newLineCount[i]

				} else {
					// Nothing new for this line
					deltaLineCount[i] = 0

				}
				oldLineCount[i] += newLineCount[i]
			}
		} else {
			// This is the first time the file is hit
			coverage.Map[newFilename] = newLineCount
			deltaMap[newFilename] = newLineCount
		}
	}
	return deltaMap
}

func calculateCoveragePercentage(coverage map[string][]int) float64 {
	runnableLines := 0
	linesRun := 0
	for _, lineCount := range coverage {
		for _, line := range lineCount {
			// This line is unreachable
			if line < 0 {
				continue
			}
			runnableLines++
			if line > 0 {
				linesRun++
			}
		}
	}
	return float64(linesRun) / float64(runnableLines) * 100
}

// Read reads from the coverage file into a map[string][]*int
// and converts to a  map[string][]int.
// Note: Coverage is a json mapping of filename (string) to an array of int
// pointers, where each element of the array indicates the line and the
// amount of coverage. Nil indicates the line is unreachable
// (i.e. whitespace, comment).  Working with pointers is annoying so replace
// with -1
func (coverage *Coverage) Read() (map[string][]int, error) {
	// Read the file
	var dst map[string][]*int
	err := util.UnmarshalFile(coverage.FilePath, &dst)
	if err != nil {
		return nil, err
	}

	// Convert from map[string][]*int to map[string][]int
	sanitized := map[string][]int{}
	for filename, lines := range dst {
		sanitizedLines := make([]int, len(lines))
		for i, line := range lines {
			if line == nil {
				sanitizedLines[i] = -1
			} else {
				sanitizedLines[i] = *line
			}
		}
		sanitized[filename] = sanitizedLines
	}
	return sanitized, nil
}

// Update reads from the coverage file and updates delta and cumulative
// coverage values
func (coverage *Coverage) Update() error {
	// Read coverage
	newCov, err := coverage.Read()
	if err != nil {
		return err
	}
	// Update coverage map
	deltaMap := coverage.updateMap(newCov)
	// Calculate the increase in coverage from the most recent request
	coverage.Delta = calculateCoveragePercentage(deltaMap)
	// Calculate the increase in coverage cumulatively
	coverage.Cumulative = calculateCoveragePercentage(coverage.Map)
	return nil
}
