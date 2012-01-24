package visor

import (
	"errors"
)

var (
	ErrAppConflict = errors.New("app is registered")
)

type Error struct{}
