// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"errors"
	"fmt"
	cp "github.com/soundcloud/cotterpin"
)

var (
	ErrConflict        = errors.New("object already exists")
	ErrInsClaimed      = errors.New("instance is already claimed")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrInvalidState    = errors.New("invalid state")
	ErrInvalidFile     = errors.New("invalid file")
	ErrBadPtyName      = errors.New("invalid proc type name: only alphanumeric chars allowed")
	ErrUnauthorized    = errors.New("operation is not permitted")
	ErrNotFound        = errors.New("object not found")
)

type Error struct {
	Err     error
	Message string
}

func NewError(err error, msg string) *Error {
	return &Error{err, msg}
}

func errorf(err error, format string, args ...interface{}) *Error {
	return &Error{err, fmt.Sprintf(format, args...)}
}

func (e *Error) Error() string {
	return e.Message
}

func IsErrConflict(e error) bool {
	return e == ErrConflict
}

func IsErrUnauthorized(e error) bool {
	return e == ErrUnauthorized
}

func IsErrNotFound(e error) bool {
	switch e.(type) {
	case *cp.Error:
		return e.(*cp.Error).Err == cp.ErrNoEnt
	case *Error:
		return e.(*Error).Err == ErrNotFound
	}
	return false
}

func IsErrInsClaimed(e error) bool {
	return e.(*Error).Err == ErrInsClaimed
}

func IsErrInvalidState(e error) bool {
	return e == ErrInvalidState
}

func IsErrInvalidFile(e error) bool {
	return e.(*Error).Err == ErrInvalidFile
}

func IsErrInvalidArgument(e error) bool {
	return e.(*Error).Err == ErrInvalidArgument
}
