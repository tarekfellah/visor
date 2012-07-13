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
	ErrUnauthorized = errors.New("operation is not permitted")
	ErrInvalidState = errors.New("invalid state")
)

type Error struct{}
