package visor

import (
	"errors"
)

var (
	ErrKeyConflict   = errors.New("key is already set")
	ErrKeyNotFound   = errors.New("key not found")
	ErrTicketClaimed = errors.New("ticket already claimed")
	ErrUnauthorized  = errors.New("operation is not permitted")
	ErrInvalidState  = errors.New("invalid state")
)

type Error struct{}
