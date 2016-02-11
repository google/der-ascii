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

import "github.com/google/der-ascii/lib"

func appendBase128(dst []byte, value uint32) []byte {
	// Special-case: zero is encoded with one, not zero bytes.
	if value == 0 {
		return append(dst, 0)
	}
	// Count how many bytes are needed.
	var l int
	for n := value; n != 0; n >>= 7 {
		l++
	}
	for ; l > 0; l-- {
		b := byte(value>>uint(7*(l-1))) & 0x7f
		if l > 1 {
			b |= 0x80
		}
		dst = append(dst, b)
	}
	return dst
}

// appendTag marshals the given tag and appends the result to dst, returning the
// updated slice.
func appendTag(dst []byte, tag lib.Tag) []byte {
	b := byte(tag.Class)
	if tag.Constructed {
		b |= 0x20
	}
	if tag.Number < 0x1f {
		// Low-tag-number form.
		b |= byte(tag.Number)
		return append(dst, b)
	}

	// High-tag-number form.
	b |= 0x1f
	dst = append(dst, b)
	return appendBase128(dst, tag.Number)
}

// appendLength marshals the given length in DER and appends the result to dst,
// returning the updated slice.
func appendLength(dst []byte, length int) []byte {
	if length < 0x80 {
		// Short-form length.
		return append(dst, byte(length))
	}

	// Long-form length. Count how many bytes are needed.
	var l byte
	for n := length; n != 0; n >>= 8 {
		l++
	}
	dst = append(dst, 0x80|l)
	for ; l > 0; l-- {
		dst = append(dst, byte(length>>uint(8*(l-1))))
	}
	return dst
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
