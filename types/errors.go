package types

import "errors"

var (
	ErrExist              = errors.New("the resource already exists")
	ErrNotExist           = errors.New("the resource does not exist")
	ErrIsDirectory        = errors.New("file is directory")
	ErrIsNotDirectory     = errors.New("file is not a directory")
	ErrEmptyRequest       = errors.New("request body is empty")
	ErrEmptyPassword      = errors.New("password is empty")
	ErrEmptyUsername      = errors.New("username is empty")
	ErrEmptyScope         = errors.New("scope is empty")
	ErrWrongDataType      = errors.New("wrong data type")
	ErrInvalidUpdateField = errors.New("invalid field to update")
	ErrInvalidOption      = errors.New("invalid option")
	ErrPathIsRel          = errors.New("path is relative")
	ErrNoPermission       = errors.New("permission denied")
	ErrInvalidAuthMethod  = errors.New("invalid auth method")
)
