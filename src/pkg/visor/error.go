package visor

import (
	"errors"
)

var (
	ErrAppConflict = errors.New("app is registered")
	ErrKeyNotFound = errors.New("environment key not found")
)

type Error struct{}
