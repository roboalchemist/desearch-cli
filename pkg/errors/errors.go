package errors

import "fmt"

// SystemError is a sentinel error type for system-level failures that should
// result in exit code 3. These include config file permission errors,
// network timeouts, and other infrastructure-level problems (as opposed to
// user errors which return exit code 1 or 2).
type SystemError struct {
	Msg string
}

func (e *SystemError) Error() string {
	return e.Msg
}

// WrapSystem returns an error wrapping msg with a SystemError sentinel.
// Use IsSystem to test whether an error (or any error in its chain) is a SystemError.
func Wrap(msg string) error {
	return &SystemError{Msg: msg}
}

// WrapF returns an error wrapping fmt.Sprintf(format, args...) with a SystemError sentinel.
func WrapF(format string, args ...interface{}) error {
	return &SystemError{Msg: fmt.Sprintf(format, args...)}
}

// IsSystem reports whether err (or any error in its chain) is a SystemError.
func IsSystem(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*SystemError)
	return ok
}
