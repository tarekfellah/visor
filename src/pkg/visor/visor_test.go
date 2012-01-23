package visor

import (
	"testing"
)

func TestDial(t *testing.T) {
	_, err := Dial(DEFAULT_ADDR)
	if err != nil {
		t.Error(err)
	}
}
