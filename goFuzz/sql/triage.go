package sql

// ValidateSQL checks for sql injection.
// TODO: if discovered, repeat params but modify to have special characters
func (queries QueryMetas) ValidateSQL() {
}

// Triage triages the provided records for any hints, errors, etc and logs them
// TODO: eventually use this info to information param mutation/flag bad behavior
func (queries QueryMetas) Triage() {
}

// TriageQueries triages the most recent sql queries
func TriageQueries(params []string) {
	// Read queries
	// Check if params showed up
	// Check for sql injections
	// Check for hints, errors, etc from postgres

}
