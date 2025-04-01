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

import "github.com/google/der-ascii/internal"

func parseBase128(bytes []byte) (ret uint32, lengthOverride int, rest []byte, ok bool) {
	rest = bytes
	if len(rest) == 0 {
		return
	}

	isMinimal := rest[0] != 0x80
	for {
		if len(rest) == 0 || (ret<<7)>>7 != ret {
			// Input too small or overflow.
			return
		}
		b := rest[0]
		ret <<= 7
		ret |= uint32(b & 0x7f)
		rest = rest[1:]
		if b&0x80 == 0 {
			ok = true
			if !isMinimal {
				lengthOverride = len(bytes) - len(rest)
			}
			return
		}
	}
}

// parseTag parses a tag from b, returning the resulting tag and the remainder
// of the slice. On parse failure, ok is returned as false and rest is
// unchanged.
func parseTag(bytes []byte) (tag internal.Tag, rest []byte, ok bool) {
	rest = bytes

	// Consume the first byte.
	if len(rest) == 0 {
		return
	}
	b := rest[0]
	rest = rest[1:]

	class := internal.Class(b & 0xc0)
	number := uint32(b & 0x1f)
	constructed := b&0x20 != 0
	if number < 0x1f {
		// Low-tag-number form.
		tag = internal.Tag{class, number, constructed, 0}
		ok = true
		return
	}

	n, lengthOverride, rest, base128Ok := parseBase128(rest)
	if !base128Ok {
		// Parse error.
		rest = bytes
		return
	}
	if n < 0x1f {
		// Non-minimal encoding.
		lengthOverride = len(bytes) - len(rest) - 1
	}
	number = n

	tag = internal.Tag{class, number, constructed, lengthOverride}
	ok = true
	return
}

type element struct {
	tag              internal.Tag
	body             []byte
	indefinite       bool
	longFormOverride int
}

// parseTagAndLength parses a tag and length pair from bytes. It is split out
// of parseElement so tests can distinguish failing to parse a length from the
// rest of the body.
func parseTagAndLength(bytes []byte) (elem element, length int, rest []byte, ok bool) {
	rest = bytes

	// Parse the tag.
	var tagOk bool
	elem.tag, bytes, tagOk = parseTag(bytes)
	if !tagOk {
		return
	}

	// Parse the length.
	if len(bytes) == 0 {
		return
	}
	b := bytes[0]
	bytes = bytes[1:]
	if b < 0x80 {
		// Short form length.
		length = int(b)
	} else if b == 0x80 {
		if !elem.tag.Constructed {
			return // Indefinite-length elements must be constructed.
		}
		elem.indefinite = true
	} else {
		// Long form length.
		b &= 0x7f
		if int(b) > len(bytes) {
			return // Not enough room.
		}
		for i := 0; i < int(b); i++ {
			if length >= 1<<23 {
				return // Overflow.
			}
			length <<= 8
			length |= int(bytes[i])
		}
		if bytes[0] == 0 || length < 0x80 {
			elem.longFormOverride = int(b) // Non-minimal length.
		}
		bytes = bytes[b:]
	}
	ok = true
	rest = bytes
	return
}

// parseElement parses an element from bytes. If the element is
// indefinite-length body is left as nil and instead indefinite is set to true.
// Note this function will treat a BER EOC as an empty element with tag number
// zero. EOC detection must be handled externally.
func parseElement(bytes []byte) (elem element, rest []byte, ok bool) {
	rest = bytes
	var length int
	elem, length, bytes, ok = parseTagAndLength(bytes)
	if !ok {
		return
	}

	if !elem.indefinite {
		if length > len(bytes) {
			ok = false
			return
		}
		elem.body = bytes[:length]
		bytes = bytes[length:]
	}
	rest = bytes
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

// decodeObjectIdentifier decodes bytes as the contents of a DER OBJECT IDENTIFIER. It
// returns the value on success and false otherwise.
func decodeObjectIdentifier(bytes []byte) (oid []uint32, ok bool) {
	// Reserve a space as the first component is split.
	oid = []uint32{0}

	// Decode each component.
	for len(bytes) != 0 {
		var c uint32
		var lengthOverride int
		c, lengthOverride, bytes, ok = parseBase128(bytes)
		if !ok || lengthOverride != 0 {
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

// decodeRelativeOID decodes bytes as the contents of a DER RELATIVE-OID. It
// returns the value on success and false otherwise.
func decodeRelativeOID(bytes []byte) (oid []uint32, ok bool) {
	// Decode each component.
	for len(bytes) != 0 {
		var c uint32
		var lengthOverride int
		c, lengthOverride, bytes, ok = parseBase128(bytes)
		if !ok || lengthOverride != 0 {
			return nil, false
		}
		oid = append(oid, c)
	}

	// Relative OIDs must have at least one component.
	if len(oid) < 1 {
		return nil, false
	}

	return oid, true
}
