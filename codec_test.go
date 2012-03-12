package visor

import (
	"testing"
)

func TestListCodecEncoding(t *testing.T) {
	codec := new(ListCodec)
	encoded, _ := codec.Encode([]string{"a", "bb", "ccc"})
	if string(encoded) != "a bb ccc" {
		t.Errorf("expected '%s' got '%s'", []byte("a bb ccc"), encoded)
	}
	return
}

func TestListCodecDecoding(t *testing.T) {
	codec := new(ListCodec)
	decoded, _ := codec.Decode([]byte("xxx yy z"))
	d := decoded.([]string)
	if (d[0] != "xxx") || (d[1] != "yy") || (d[2] != "z") {
		t.Error("Couldn't decode internal list representation")
	}
	return
}
