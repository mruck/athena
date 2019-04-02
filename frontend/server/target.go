package server

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
)

// TargetDB containers metadata about the database we are instrumenting
type TargetDB struct {
	User     *string
	Password *string
	Host     *string
	Port     *int
	Name     *string
}

// Target is the expected form of user input
type Target struct {
	// Name of the target application
	Name *string
	// Port the target app is running on
	Port       *int
	Db         *TargetDB
	Containers []v1.Container
}

// ValidateTarget checks user provided input.  We make all values in Target and TargetDB
// be pointers because that's the only way to validate them
func ValidateTarget(target *Target) error {
	if target.Name == nil {
		return fmt.Errorf("error: must specify target name")
	}
	if target.Port == nil {
		return fmt.Errorf("error: must specify target port")
	}
	if target.Containers == nil {
		return fmt.Errorf("error: must specify []v1.Container")
	}
	if target.Db == nil {
		return fmt.Errorf("error: must specify db metadata")
	}
	if target.Db.User == nil {
		return fmt.Errorf("error: must specify db user")
	}
	if target.Db.Password == nil {
		return fmt.Errorf("error: must specify db password")
	}
	if target.Db.Host == nil {
		return fmt.Errorf("error: must specify db host")
	}
	if target.Db.Port == nil {
		return fmt.Errorf("error: must specify db port")
	}
	if target.Db.Name == nil {
		return fmt.Errorf("error: must specify db name")
	}
	// Ensure the list of containers provided includes a container named "target".
	// This is where rails-fork lives, and also the target application source code
	// for instrumenting.
	if GetTargetContainer(target.Containers) == nil {
		return fmt.Errorf("Please provide a container named \"target\"")
	}
	return nil
}
