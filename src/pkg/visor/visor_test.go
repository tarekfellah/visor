package visor

import (
	"testing"
)

func TestDialWithDefaultAddr(t *testing.T) {
	_, err := Dial(DEFAULT_ADDR)
	if err != nil {
		t.Error(err)
	}
}

func TestDialWithInvalidAddr(t *testing.T) {
	_, err := Dial("foo.bar:123:876")
	if err == nil {
		t.Error("Dialed with invalid addr")
	}
}
