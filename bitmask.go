package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
)

type Bitmask uint64

func (f Bitmask) HasFlag(flag Bitmask) bool { return f&flag != 0 }
func (f *Bitmask) AddFlag(flag Bitmask)     { *f |= flag }
func (f *Bitmask) ClearFlag(flag Bitmask)   { *f &= ^flag }
func (f *Bitmask) ToggleFlag(flag Bitmask)  { *f ^= flag }

// Implementing json.Marshaler interface for Bitmask
func (b Bitmask) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, b)
	if err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return json.Marshal(encoded)
}

// Implementing json.Unmarshaler interface for Bitmask
func (b *Bitmask) UnmarshalJSON(data []byte) error {
	var encoded string
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}
	buf := bytes.NewReader(decoded)
	err = binary.Read(buf, binary.BigEndian, b)
	if err != nil {
		return err
	}
	return nil
}
