package util

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mruck/athena/lib/log"
	"github.com/pkg/errors"
)

// RandInt returns a truncated uuid
func RandInt() uint32 {
	uid := uuid.New()
	return uid.ID()
}

// RandString returns a stringified uuid
func RandString() string {
	uid := uuid.New()
	return uid.String()
}

// RandBool returns true or false
// TODO: make this actually random!
func RandBool() bool {
	return true
}

// RandDecimal returns true or false
func RandDecimal() float32 {
	return float32(RandInt()) / 100
}

// Rand returns a random object of type typ.
// Returns a random string if the data type doesn't match
func Rand(dataType string) interface{} {
	// TODO: use a rng seeded with 0 for reproducability?
	if dataType == "string" {
		return RandString()
	}
	if dataType == "integer" {
		return RandInt()
	}
	if dataType == "number" {
		return RandInt()
	}
	if dataType == "boolean" {
		return RandBool()
	}
	if dataType == "decimal" {
		return RandDecimal()
	}
	err := fmt.Errorf("util.Rand() called on unsupport data type: %s", dataType)
	log.Errorf("%+v", errors.WithStack(err))
	return RandString()
}
