package util

import (
	fuzz "github.com/google/gofuzz"
	"github.com/google/uuid"
)

// RandString returns a stringified uuid.
// This data should use a much more normal encoding
// than go fuzz
func RandString() string {
	uid := uuid.New()
	return uid.String()[:4]
}

// Rand returns a random object of type typ.
// Returns a random string if the data type doesn't match
func Rand(dataType string) interface{} {
	f := fuzz.New()
	switch dataType {
	case "integer":
		fallthrough
	case "number":
		var val int
		f.Fuzz(&val)
		return val
	case "boolean":
		var val bool
		f.Fuzz(&val)
		return val
	case "decimal":
		var val float32
		f.Fuzz(&val)
		return val
	case "string":
		fallthrough
	default:
		return RandString()
		//var val string
		//f.Fuzz(&val)
		//return val
	}
}
