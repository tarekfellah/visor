// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// A Codec represents a protocol for encoding and
// decoding file values in the coordinator.
type codec interface {
	Encode(input interface{}) ([]byte, error)
	Decode(input []byte) (interface{}, error)
}

// ByteCodec is a transparent Codec which doesn't
// perform any serialization or deserialization.
type byteCodec struct{}

func (*byteCodec) Encode(input interface{}) ([]byte, error) {
	return input.([]byte), nil
}
func (*byteCodec) Decode(input []byte) (interface{}, error) {
	return input, nil
}

// StringCodec is a Codec which converts data to and from the Go *string* type.
type stringCodec struct{}

func (*stringCodec) Encode(input interface{}) (output []byte, err error) {
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
func (*stringCodec) Decode(input []byte) (interface{}, error) {
	return string(input), nil
}

type jsonCodec struct {
	// If non-nil, unmarshal into this receiver. This is needed to decode JSON
	// values into Go structs.
	decodedVal interface{}
}

func (*jsonCodec) Encode(input interface{}) ([]byte, error) {
	return json.Marshal(input)
}
func (c *jsonCodec) Decode(input []byte) (val interface{}, err error) {
	// If the user specified an explicit receiver object, unmarshal into that.
	if c.decodedVal != nil {
		val = c.decodedVal
	}
	err = json.Unmarshal(input, &val)
	return
}

type intCodec struct{}

func (*intCodec) Encode(input interface{}) ([]byte, error) {
	return []byte(strconv.Itoa(input.(int))), nil
}
func (*intCodec) Decode(input []byte) (interface{}, error) {
	return strconv.Atoi(string(input))
}

type listCodec struct{}

func (*listCodec) Encode(input interface{}) ([]byte, error) {
	return []byte(strings.TrimSpace(strings.Join(input.([]string), " "))), nil
}
func (*listCodec) Decode(input []byte) (interface{}, error) {
	return strings.Fields(string(input)), nil
}
