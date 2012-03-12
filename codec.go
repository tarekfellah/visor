package visor

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// A Codec represents a protocol for encoding and
// decoding file values in the coordinator.
type Codec interface {
	Encode(input interface{}) ([]byte, error)
	Decode(input []byte) (interface{}, error)
}

// ByteCodec is a transparent Codec which doesn't
// perform any serialization or deserialization.
type ByteCodec struct{}

func (*ByteCodec) Encode(input interface{}) ([]byte, error) {
	return input.([]byte), nil
}
func (*ByteCodec) Decode(input []byte) (interface{}, error) {
	return input, nil
}

// StringCodec is a Codec which converts data to and from the Go *string* type.
type StringCodec struct{}

func (*StringCodec) Encode(input interface{}) (output []byte, err error) {
	switch i := input.(type) {
	case string:
		output = []byte(i)
	case []byte: // TODO: do we want allow bytes?
		output = i
	default:
		err = errors.New("expected string or []byte input")
	}
	return
}
func (*StringCodec) Decode(input []byte) (interface{}, error) {
	return string(input), nil
}

type JSONCodec struct{}

func (*JSONCodec) Encode(input interface{}) ([]byte, error) {
	return json.Marshal(input)
}
func (*JSONCodec) Decode(input []byte) (val interface{}, err error) {
	err = json.Unmarshal(input, &val)
	return
}

type IntCodec struct{}

func (*IntCodec) Encode(input interface{}) ([]byte, error) {
	return []byte(strconv.Itoa(input.(int))), nil
}
func (*IntCodec) Decode(input []byte) (interface{}, error) {
	return strconv.Atoi(string(input))
}

type ListCodec struct{}

func (*ListCodec) Encode(input interface{}) ([]byte, error) {
	return []byte(strings.Join(input.([]string), " ")), nil
}
func (*ListCodec) Decode(input []byte) (interface{}, error) {
	return strings.Split(string(input), " "), nil
}
