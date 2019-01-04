package lib

import "errors"

var (
	ErrExist             = errors.New("the resource already exists")
	ErrNotExist          = errors.New("the resource does not exist")
	ErrIsDirectory       = errors.New("file is directory")
	ErrEmptyPassword     = errors.New("password is empty")
	ErrEmptyUsername     = errors.New("username is empty")
	ErrInvalidOption     = errors.New("invalid option")
	ErrPathIsRel         = errors.New("path is relative")
	ErrNoPermission      = errors.New("permission denied")
	ErrInvalidAuthMethod = errors.New("invalid auth method")
	ErrEmptyKey          = errors.New("empty key")
	ErrInvalidDataType   = errors.New("invalid data type")
)
