// Copyright 2015 The DER ASCII Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"fmt"

	"github.com/google/der-ascii/internal"
)

func appendBase128(dst []byte, value uint32) []byte {
	dst, err := appendBase128WithLength(dst, value, 0)
	if err != nil {
		// Only a length override can fail.
		panic(err)
	}
	return dst
}

func appendBase128WithLength(dst []byte, value uint32, length int) ([]byte, error) {
	// Count how many bytes are needed.
	var l int
	for n := value; n != 0; n >>= 7 {
		l++
	}
	// Special-case: zero is encoded with one, not zero bytes.
	if value == 0 {
		l = 1
	}
	// Apply the length override.
	if length != 0 {
		if length < l {
			return nil, fmt.Errorf("length override of %d is too small, need at least %d bytes", length, l)
		}
		l = length
	}
	for ; l > 0; l-- {
		b := byte(value>>uint(7*(l-1))) & 0x7f
		if l > 1 {
			b |= 0x80
		}
		dst = append(dst, b)
	}
	return dst, nil
}

// appendTag marshals the given tag and appends the result to dst, returning the
// updated slice.
func appendTag(dst []byte, tag internal.Tag) ([]byte, error) {
	b := byte(tag.Class)
	if tag.Constructed {
		b |= 0x20
	}
	if tag.Number < 0x1f && tag.LongFormOverride == 0 {
		// Low-tag-number form.
		b |= byte(tag.Number)
		return append(dst, b), nil
	}

	// High-tag-number form.
	b |= 0x1f
	dst = append(dst, b)
	return appendBase128WithLength(dst, tag.Number, tag.LongFormOverride)
}

// appendLength marshals the given length in DER and appends the result to dst,
// returning the updated slice.
func appendLength(dst []byte, length, lengthLength int) ([]byte, error) {
	if length < 0x80 && lengthLength == 0 {
		// Short-form length.
		return append(dst, byte(length)), nil
	}

	// Long-form length. Count how many bytes are needed.
	var l int
	for n := length; n != 0; n >>= 8 {
		l++
	}
	if lengthLength != 0 {
		if lengthLength > 127 {
			return nil, errors.New("length override too large")
		}
		if lengthLength < l {
			return nil, fmt.Errorf("length override of %d too small, need at least %d bytes", lengthLength, l)
		}
		l = lengthLength
	}
	dst = append(dst, 0x80|byte(l))
	for ; l > 0; l-- {
		dst = append(dst, byte(length>>uint(8*(l-1))))
	}
	return dst, nil
}

// appendInteger marshals the given value as the contents of a DER INTEGER and
// appends the result to dst, returning the updated slice.
func appendInteger(dst []byte, value int64) []byte {
	// Count how many bytes are needed.
	l := 1
	for n := value; n > 0x7f || n < (0x80-0x100); n >>= 8 {
		l++
	}

	for ; l > 0; l-- {
		dst = append(dst, byte(value>>uint(8*(l-1))))
	}
	return dst
}

func appendObjectIdentifier(dst []byte, value []uint32) ([]byte, bool) {
	// Validate the input before anything is written.
	if len(value) < 2 || value[0] > 2 || (value[0] < 2 && value[1] > 39) {
		return dst, false
	}
	if value[0]*40+value[1] < value[1] {
		return dst, false
	}

	dst = appendBase128(dst, value[0]*40+value[1])
	for _, v := range value[2:] {
		dst = appendBase128(dst, v)
	}
	return dst, true
}

func appendRelativeOID(dst []byte, value []uint32) []byte {
	for _, v := range value {
		dst = appendBase128(dst, v)
	}
	return dst
}
