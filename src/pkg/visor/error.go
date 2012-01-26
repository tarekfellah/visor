package visor

import (
	"errors"
)

var (
	ErrKeyConflict = errors.New("key is already set")
	ErrKeyNotFound = errors.New("key not found")
)

type Error struct{}
