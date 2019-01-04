package errors

import "errors"

var (
	ErrEmptyKey          = errors.New("empty key")
	ErrExist             = errors.New("the resource already exists")
	ErrNotExist          = errors.New("the resource does not exist")
	ErrEmptyPassword     = errors.New("password is empty")
	ErrEmptyUsername     = errors.New("username is empty")
	ErrEmptyRequest      = errors.New("empty request")
	ErrScopeIsRelative   = errors.New("scope is a relative path")
	ErrInvalidDataType   = errors.New("invalid data type")
	ErrIsDirectory       = errors.New("file is directory")
	ErrInvalidOption     = errors.New("invalid option")
	ErrInvalidAuthMethod = errors.New("invalid auth method")
)
