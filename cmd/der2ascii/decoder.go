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

func parseBase128(bytes []byte) (ret uint32, rest []byte, ok bool) {
	// The tag must be minimally-encoded, so the first byte may not be 0x80.
	if len(bytes) == 0 || bytes[0] == 0x80 {
		return 0, bytes, false
	}

	for {
		if len(bytes) == 0 || (ret<<7)>>7 != ret {
			// Input too small or overflow.
			return 0, bytes, false
		}
		b := bytes[0]
		ret <<= 7
		ret |= uint32(b & 0x7f)
		bytes = bytes[1:]
		if b&0x80 == 0 {
			return ret, bytes, true
		}
	}
}

// parseTag parses a tag from b, returning the resulting tag and the remainder
// of the slice. On parse failure, ok is returned as false and rest is
// unchanged.
func parseTag(bytes []byte) (tag lib.Tag, rest []byte, ok bool) {
	rest = bytes

	// Consume the first byte. Reject EOC.
	if len(rest) == 0 || rest[0] == 0 {
		return
	}
	b := rest[0]
	rest = rest[1:]

	class := lib.Class(b & 0xc0)
	number := uint32(b & 0x1f)
	constructed := b&0x20 != 0
	if number < 0x1f {
		// Low-tag-number form.
		tag = lib.Tag{class, number, constructed}
		ok = true
		return
	}

	n, rest, base128Ok := parseBase128(rest)
	if !base128Ok || n < 0x1f {
		// Parse error or non-minimal encoding.
		rest = bytes
		return
	}
	number = n

	tag = lib.Tag{class, number, constructed}
	ok = true
	return
}

// parseTagAndLength parses a tag and length pair from bytes. If the resulting
// length is indefinite, it sets indefinite to true.
func parseTagAndLength(bytes []byte) (tag lib.Tag, length int, indefinite bool, rest []byte, ok bool) {
	rest = bytes

	// Parse the tag.
	tag, rest, ok = parseTag(rest)
	if !ok {
		return lib.Tag{}, 0, false, bytes, false
	}

	// Parse the length.
	if len(rest) == 0 {
		return lib.Tag{}, 0, false, bytes, false
	}
	b := rest[0]
	rest = rest[1:]
	if b < 0x80 {
		// Short form length.
		length = int(b)
		return
	}
	if b == 0x80 {
		// Indefinite-length. Must be constructed.
		if !tag.Constructed {
			return lib.Tag{}, 0, false, bytes, false
		}
		indefinite = true
		return
	}
	// Long form length.
	b &= 0x7f
	if int(b) > len(rest) || rest[0] == 0 {
		// Not enough room or non-minimal length.
		return lib.Tag{}, 0, false, bytes, false
	}
	for i := 0; i < int(b); i++ {
		if length >= 1<<23 {
			// Overflow.
			return lib.Tag{}, 0, false, bytes, false
		}
		length <<= 8
		length |= int(rest[i])
	}
	if length < 0x80 {
		// Should have been short form.
		return lib.Tag{}, 0, false, bytes, false
	}
	rest = rest[b:]
	return
}

// parseElement parses an element from bytes. If the element is
// indefinite-length body is left as nil and instead indefinite is set to true.
func parseElement(bytes []byte) (tag lib.Tag, body []byte, indefinite bool, rest []byte, ok bool) {
	rest = bytes

	tag, length, indefinite, rest, ok := parseTagAndLength(rest)
	if !ok || length > len(rest) {
		return lib.Tag{}, nil, false, bytes, false
	}

	body = rest[:length]
	rest = rest[length:]
	return
}

// decodeInteger decodes bytes as the contents of a DER INTEGER. It returns the
// value on success and false otherwise.
func decodeInteger(bytes []byte) (int64, bool) {
	if len(bytes) == 0 {
		return 0, false
	}

	// Reject non-minimal encodings.
	if len(bytes) > 1 && (bytes[0] == 0 || bytes[0] == 0xff) && bytes[0]&0x80 == bytes[1]&0x80 {
		return 0, false
	}

	val := int64(bytes[0])
	if val&0x80 != 0 {
		val -= 256
	}
	for _, v := range bytes[1:] {
		if (val<<8)>>8 != val {
			return 0, false
		}
		val <<= 8
		val |= int64(v)
	}
	return val, true
}

// decodeInteger decodes bytes as the contents of a DER OBJECT IDENTIFIER. It
// returns the value on success and false otherwise.
func decodeObjectIdentifier(bytes []byte) (oid []uint32, ok bool) {
	// Reserve a space as the first component is split.
	oid = []uint32{0}

	// Decode each component.
	for len(bytes) != 0 {
		var c uint32
		c, bytes, ok = parseBase128(bytes)
		if !ok {
			return nil, false
		}
		oid = append(oid, c)
	}

	// OIDs must have at least two components.
	if len(oid) < 2 {
		return nil, false
	}

	// Adjust the first component.
	if oid[1] >= 80 {
		oid[0] = 2
		oid[1] -= 80
	} else if oid[1] >= 40 {
		oid[0] = 1
		oid[1] -= 40
	}

	return oid, true
}
