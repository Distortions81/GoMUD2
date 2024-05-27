package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

const BITMASK_STRING_BASE = 10

type Bitmask uint64

func (f Bitmask) hasFlag(flag Bitmask) bool { return f&flag != 0 }
func (f *Bitmask) addFlag(flag Bitmask)     { *f |= flag }
func (f *Bitmask) clearFlag(flag Bitmask)   { *f &= ^flag }
func (f *Bitmask) toggleFlag(flag Bitmask)  { *f ^= flag }

const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

func (b Bitmask) MarshalJSON() ([]byte, error) {
	base64String := uint64ToBase64(uint64(b))
	return json.Marshal(base64String)
}

func (b *Bitmask) UnmarshalJSON(data []byte) error {
	var base64String string
	if err := json.Unmarshal(data, &base64String); err != nil {
		return err
	}

	num, err := base64ToUint64(base64String)
	if err != nil {
		return err
	}

	*b = Bitmask(num)
	return nil
}

func uint64ToBase64(num uint64) string {
	var result string
	for num > 0 {
		index := num & 0x3f // Mask to get the last 6 bits
		result = string(base64Chars[index]) + result
		num >>= 6 // Shift right by 6 bits
	}
	return result
}

func base64ToUint64(base64String string) (uint64, error) {
	var num uint64
	for i := 0; i < len(base64String); i++ {
		char := base64String[i]
		index := strings.IndexByte(base64Chars, char)
		if index == -1 {
			return 0, fmt.Errorf("invalid base64 character: %c", char)
		}
		num = (num << 6) | uint64(index) // Shift left by 6 bits and add index
	}
	return num, nil
}
