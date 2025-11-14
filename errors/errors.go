package errors

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

const (
	ExitCodeSigTerm = 128 + int(syscall.SIGTERM)
	ExitCodeSighup  = 128 + int(syscall.SIGHUP)
	ExitCodeSigint  = 128 + int(syscall.SIGINT)
	ExitCodeSigquit = 128 + int(syscall.SIGQUIT)
)

var (
	ErrEmptyKey             = errors.New("empty key")
	ErrExist                = errors.New("the resource already exists")
	ErrNotExist             = errors.New("the resource does not exist")
	ErrEmptyPassword        = errors.New("password is empty")
	ErrEasyPassword         = errors.New("password is too easy")
	ErrEmptyUsername        = errors.New("username is empty")
	ErrEmptyRequest         = errors.New("empty request")
	ErrScopeIsRelative      = errors.New("scope is a relative path")
	ErrInvalidDataType      = errors.New("invalid data type")
	ErrIsDirectory          = errors.New("file is directory")
	ErrInvalidOption        = errors.New("invalid option")
	ErrInvalidAuthMethod    = errors.New("invalid auth method")
	ErrPermissionDenied     = errors.New("permission denied")
	ErrInvalidRequestParams = errors.New("invalid request params")
	ErrSourceIsParent       = errors.New("source is parent")
	ErrRootUserDeletion     = errors.New("user with id 1 can't be deleted")
	ErrSigTerm              = errors.New("exit on signal: sigterm")
	ErrSighup               = errors.New("exit on signal: sighup")
	ErrSigint               = errors.New("exit on signal: sigint")
	ErrSigquit              = errors.New("exit on signal: sigquit")
)

type ErrShortPassword struct {
	MinimumLength uint
}

func (e ErrShortPassword) Error() string {
	return fmt.Sprintf("password is too short, minimum length is %d", e.MinimumLength)
}

// GetExitCode returns the exit code for a given error.
func GetExitCode(err error) int {
	if err == nil {
		return 0
	}

	exitCodeMap := map[error]int{
		ErrSigTerm: ExitCodeSigTerm,
		ErrSighup:  ExitCodeSighup,
		ErrSigint:  ExitCodeSigint,
		ErrSigquit: ExitCodeSigquit,
	}

	for e, code := range exitCodeMap {
		if errors.Is(err, e) {
			return code
		}
	}

	if exitErr, ok := err.(interface{ ExitCode() int }); ok {
		return exitErr.ExitCode()
	}

	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return 1
	}

	var syscallErr *os.SyscallError
	if errors.As(err, &syscallErr) {
		return 1
	}

	var errno syscall.Errno
	if errors.As(err, &errno) {
		return 1
	}

	return 1
}
