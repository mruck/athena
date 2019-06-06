package log

// Check how log lib prints stack trace.  Looks like we stil need to do
// errors.WithStack

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
)

func stack2() error {
	return errors.WithStack(fmt.Errorf("This is an error"))
	//return fmt.Errorf("This is an error")
}

func stack1() error {
	return stack2()
}

func TestStackTrace(t *testing.T) {
	Errorf("Something went wrong: %+v\n", stack1())
}

func TestStackTraceInfo(t *testing.T) {
	Infof("%+v\n", stack1())
}

func TestLogToFile(t *testing.T) {
	Info("hello")
	Error("this is an error")
	Error("this is an error2")
}

//func TestPanic(t *testing.T) {
//	Panic("printing a fatal error")
//	Info("Do we continue execution")
//}

//func TestFatal(t *testing.T) {
//	Fatal("printing a fatal error")
//	Info("Do we continue execution")
//}
