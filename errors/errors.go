package fberrors

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyKey                  = errors.New("empty key")
	ErrExist                     = errors.New("the resource already exists")
	ErrNotExist                  = errors.New("the resource does not exist")
	ErrEmptyPassword             = errors.New("password is empty")
	ErrEasyPassword              = errors.New("password is too easy")
	ErrEmptyUsername             = errors.New("username is empty")
	ErrEmptyRequest              = errors.New("empty request")
	ErrScopeIsRelative           = errors.New("scope is a relative path")
	ErrInvalidDataType           = errors.New("invalid data type")
	ErrIsDirectory               = errors.New("file is directory")
	ErrInvalidOption             = errors.New("invalid option")
	ErrInvalidAuthMethod         = errors.New("invalid auth method")
	ErrPermissionDenied          = errors.New("permission denied")
	ErrInvalidRequestParams      = errors.New("invalid request params")
	ErrSourceIsParent            = errors.New("source is parent")
	ErrRootUserDeletion          = errors.New("user with id 1 can't be deleted")
	ErrCurrentPasswordIncorrect  = errors.New("the current password is incorrect")
	ErrZipFileIsTooLarge         = errors.New("the zip file is too large")
	ErrCompressionRateIsTooLarge = errors.New("one of the files has too high a decompression rate")
	ErrInvalidZipFilePath        = errors.New("invalid path in some files of the zip archive")
	ErrUncompressSizeIsTooLarge  = errors.New("one of the files has too high a decompression size")
	ErrInvalidZipEntry           = errors.New("some files are invalid in zip archive")
)

type ErrShortPassword struct {
	MinimumLength uint
}

func (e ErrShortPassword) Error() string {
	return fmt.Sprintf("password is too short, minimum length is %d", e.MinimumLength)
}
