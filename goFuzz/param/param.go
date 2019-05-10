package param

// Param keeps track of metadata surrounding a parameter
type Param struct {
	Name           string
	Type           []string
	Value          string
	PreviousValues []string
	NextValues     []string
}

// New returns a new param
func New(name string) *Param {
	return &Param{Name: name}
}

// Mutate a parameter
func (param *Param) Mutate() {
	// Check for database queries

}
