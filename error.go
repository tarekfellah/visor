// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"errors"
)

var (
	ErrKeyConflict  = errors.New("key is already set")
	ErrRevMismatch  = errors.New("revision mismatch")
	ErrInsClaimed   = errors.New("instance is already claimed")
	ErrUnauthorized = errors.New("operation is not permitted")
	ErrInvalidState = errors.New("invalid state")
	ErrNoEnt        = errors.New("file not found")
	ErrBadPath      = errors.New("invalid path: only ASCII letters, numbers, '.', or '-' are allowed")
)

type Error struct {
	Err     error
	Message string
}

func NewError(err error, msg string) *Error {
	return &Error{err, msg}
}

func (e *Error) Error() string {
	return e.Message
}

func IsErrNoEnt(e error) (r bool) {
	if err, ok := e.(*Error); ok {
		r = err.Err == ErrNoEnt
	}
	return
}
