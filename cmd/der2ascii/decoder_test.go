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
	"bytes"
	"math"
	"testing"

	"github.com/google/der-ascii/lib"
)

var parseTagTests = []struct {
	in  []byte
	tag lib.Tag
	ok  bool
}{
	{[]byte{0x30}, lib.Tag{lib.ClassUniversal, 16, true}, true},
	{[]byte{0x02}, lib.Tag{lib.ClassUniversal, 2, false}, true},
	{[]byte{0x7f, 0x89, 0x52}, lib.Tag{lib.ClassApplication, 1234, true}, true},
	// Empty.
	{[]byte{}, lib.Tag{}, false},
	// Truncated high-tag-number-form.
	{[]byte{0x7f}, lib.Tag{}, false},
	{[]byte{0x7f, 0xff}, lib.Tag{}, false},
	// Should have been low-tag-number form.
	{[]byte{0x7f, 0x01}, lib.Tag{}, false},
	// Non-minimal encoding.
	{[]byte{0x7f, 0x00, 0x89, 0x52, 0x00}, lib.Tag{}, false},
	// Overflow.
	{[]byte{0xff, 0x8f, 0xff, 0xff, 0xff, 0x7f}, lib.Tag{lib.ClassPrivate, (1 << 32) - 1, true}, true},
	{[]byte{0xff, 0x9f, 0xff, 0xff, 0xff, 0x7f}, lib.Tag{}, false},
	// EOC.
	{[]byte{0x00}, lib.Tag{}, false},
}

func TestParseTag(t *testing.T) {
	for i, tt := range parseTagTests {
		tag, rest, ok := parseTag(tt.in)
		if !tt.ok {
			if ok {
				t.Errorf("%d. parseTag(%v) unexpectedly succeeded.", i, tt.in)
			} else if !bytes.Equal(rest, tt.in) {
				t.Errorf("%d. parseTag(%v) did not preserve input.", i, tt.in)
			}
		} else {
			if !ok {
				t.Errorf("%d. parseTag(%v) unexpectedly failed.", i, tt.in)
			} else if tag != tt.tag || len(rest) != 0 {
				t.Errorf("%d. parseTag(%v) = %v, %v wanted %v, [].", i, tt.in, tag, rest, tt.tag)
			}

			// Test again with trailing data.
			in := make([]byte, len(tt.in)+5)
			copy(in, tt.in)
			tag, rest, ok = parseTag(in)
			if !ok {
				t.Errorf("%d. parseTag(%v) unexpectedly failed.", i, in)
			} else if tag != tt.tag || !bytes.Equal(rest, in[len(tt.in):]) {
				t.Errorf("%d. parseTag(%v) = %v, %v wanted %v, %v.", i, in, tag, rest, tt.tag, in[len(tt.in):])
			}
		}
	}
}

var sequenceTag = lib.Tag{lib.ClassUniversal, 16, true}

var parseTagAndLengthTests = []struct {
	in         []byte
	tag        lib.Tag
	length     int
	indefinite bool
	ok         bool
}{
	// Short-form length.
	{[]byte{0x30, 0x00}, sequenceTag, 0, false, true},
	{[]byte{0x30, 0x01}, sequenceTag, 1, false, true},
	// Indefinite length.
	{[]byte{0x30, 0x80}, sequenceTag, 0, true, true},
	// Long-form length.
	{[]byte{0x30, 0x81, 0x80}, sequenceTag, 128, false, true},
	{[]byte{0x30, 0x81, 0xff}, sequenceTag, 255, false, true},
	{[]byte{0x30, 0x82, 0x01, 0x00}, sequenceTag, 256, false, true},
	// Too short.
	{[]byte{0x30}, lib.Tag{}, 0, false, false},
	{[]byte{0x30, 0x81}, lib.Tag{}, 0, false, false},
	// Non-minimal form length.
	{[]byte{0x30, 0x82, 0x00, 0xff}, lib.Tag{}, 0, false, false},
	{[]byte{0x30, 0x81, 0x1f}, lib.Tag{}, 0, false, false},
	// Overflow.
	{[]byte{0x30, 0x85, 0xff, 0xff, 0xff, 0xff, 0xff}, lib.Tag{}, 0, false, false},
	// Empty.
	{[]byte{}, lib.Tag{}, 0, false, false},
	// Primitive + indefinite length is illegal.
	{[]byte{0x02, 0x80}, lib.Tag{}, 0, false, false},
}

func TestParseTagAndLength(t *testing.T) {
	for i, tt := range parseTagAndLengthTests {
		tag, length, indefinite, rest, ok := parseTagAndLength(tt.in)
		if !tt.ok {
			if ok {
				t.Errorf("%d. parseTagAndLength(%v) unexpectedly succeeded.", i, tt.in)
			} else if !bytes.Equal(rest, tt.in) {
				t.Errorf("%d. parseTagAndLength(%v) did not preserve input.", i, tt.in)
			}
		} else {
			if !ok {
				t.Errorf("%d. parseTagAndLength(%v) unexpectedly failed.", i, tt.in)
			} else if tag != tt.tag || length != tt.length || indefinite != tt.indefinite || len(rest) != 0 {
				t.Errorf("%d. parseTagAndLength(%v) = %v, %v, %v, %v wanted %v, %v, %v, [].", i, tt.in, tag, length, indefinite, rest, tt.tag, tt.length, tt.indefinite)
			}

			// Test again with trailing data.
			in := make([]byte, len(tt.in)+5)
			copy(in, tt.in)
			tag, length, indefinite, rest, ok := parseTagAndLength(in)
			if !ok {
				t.Errorf("%d. parseTagAndLength(%v) unexpectedly failed.", i, in)
			} else if tag != tt.tag || length != tt.length || indefinite != tt.indefinite || !bytes.Equal(rest, in[len(tt.in):]) {
				t.Errorf("%d. parseTagAndLength(%v) = %v, %v, %v, %v wanted %v, %v.", i, in, tag, length, indefinite, rest, tt.tag, tt.length, tt.indefinite, in[len(tt.in):])
			}
		}
	}
}

var parseElementTests = []struct {
	in         []byte
	tag        lib.Tag
	body       []byte
	indefinite bool
	ok         bool
}{
	// Normal element.
	{[]byte{0x30, 0x00}, sequenceTag, []byte{}, false, true},
	{[]byte{0x30, 0x01, 0xaa}, sequenceTag, []byte{0xaa}, false, true},
	// Indefinite length.
	{[]byte{0x30, 0x80}, sequenceTag, nil, true, true},
	// Too short.
	{[]byte{0x30}, lib.Tag{}, []byte{}, false, false},
	{[]byte{0x30, 0x01}, lib.Tag{}, []byte{}, false, false},
	{[]byte{0x30, 0x81}, lib.Tag{}, []byte{}, false, false},
}

func TestParseElement(t *testing.T) {
	for i, tt := range parseElementTests {
		tag, body, indefinite, rest, ok := parseElement(tt.in)
		if !tt.ok {
			if ok {
				t.Errorf("%d. parseElement(%v) unexpectedly succeeded.", i, tt.in)
			} else if !bytes.Equal(rest, tt.in) {
				t.Errorf("%d. parseElement(%v) did not preserve input.", i, tt.in)
			}
		} else {
			if !ok {
				t.Errorf("%d. parseElement(%v) unexpectedly failed.", i, tt.in)
			} else if tag != tt.tag || !bytes.Equal(body, tt.body) || indefinite != tt.indefinite || len(rest) != 0 {
				t.Errorf("%d. parseElement(%v) = %v, %v, %v, %v wanted %v, %v, %v, [].", i, tt.in, tag, body, indefinite, rest, tt.tag, tt.body, tt.indefinite)
			}

			// Test again with trailing data.
			in := make([]byte, len(tt.in)+5)
			copy(in, tt.in)
			tag, body, indefinite, rest, ok := parseElement(in)
			if !ok {
				t.Errorf("%d. parseElement(%v) unexpectedly failed.", i, in)
			} else if tag != tt.tag || !bytes.Equal(body, tt.body) || indefinite != tt.indefinite || !bytes.Equal(rest, in[len(tt.in):]) {
				t.Errorf("%d. parseElement(%v) = %v, %v, %v, %v wanted %v, %v.", i, in, tag, body, indefinite, rest, tt.tag, tt.body, tt.indefinite, in[len(tt.in):])
			}
		}
	}
}

var decodeIntegerTests = []struct {
	in  []byte
	out int64
	ok  bool
}{
	// Valid encodings.
	{[]byte{0x00}, 0, true},
	{[]byte{0x01}, 1, true},
	{[]byte{0xff}, -1, true},
	{[]byte{0x7f}, 127, true},
	{[]byte{0x00, 0x80}, 128, true},
	{[]byte{0x01, 0x00}, 256, true},
	{[]byte{0x80}, -128, true},
	{[]byte{0xff, 0x7f}, -129, true},
	// Empty encoding.
	{[]byte{}, 0, false},
	// Non-minimal encodings.
	{[]byte{0x00, 0x01}, 0, false},
	{[]byte{0xff, 0xff}, 0, false},
	// Overflow tests.
	{[]byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, (1 << 63) - 1, true},
	{[]byte{0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0, false},
	{[]byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, -(1 << 63), true},
	{[]byte{0xff, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 0, false},
}

func TestDecodeInteger(t *testing.T) {
	for i, tt := range decodeIntegerTests {
		out, ok := decodeInteger(tt.in)
		if !tt.ok {
			if ok {
				t.Errorf("%d. decodeInteger(%v) unexpectedly succeeded.", i, tt.in)
			}
		} else if !ok {
			t.Errorf("%d. decodeInteger(%v) unexpectedly failed.", i, tt.in)
		} else if out != tt.out {
			t.Errorf("%d. decodeInteger(%v) = %v wanted %v.", i, tt.in, out, tt.out)
		}
	}
}

func eqUint32s(a, b []uint32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

var decodeObjectIdentifierTests = []struct {
	in  []byte
	out []uint32
	ok  bool
}{
	{[]byte{1}, []uint32{0, 1}, true},
	{[]byte{42, 3, 4, 0x7f, 0x81, 0x00, 0x81, 0x01}, []uint32{1, 2, 3, 4, 127, 128, 129}, true},
	{[]byte{81}, []uint32{2, 1}, true},
	{[]byte{0x8f, 0xff, 0xff, 0xff, 0x7f}, []uint32{2, math.MaxUint32 - 80}, true},
	// Empty.
	{[]byte{}, nil, false},
	// Incomplete component.
	{[]byte{0xff}, nil, false},
	// Overflow.
	{[]byte{0x9f, 0xff, 0xff, 0xff, 0x7f}, nil, false},
}

func TestDecodeObjectIdentifier(t *testing.T) {
	for i, tt := range decodeObjectIdentifierTests {
		out, ok := decodeObjectIdentifier(tt.in)
		if !tt.ok {
			if ok {
				t.Errorf("%d. decodeObjectIdentifier(%v) unexpectedly succeeded.", i, tt.in)
			}
		} else if !ok {
			t.Errorf("%d. decodeObjectIdentifier(%v) unexpectedly failed.", i, tt.in)
		} else if !eqUint32s(out, tt.out) {
			t.Errorf("%d. decodeObjectIdentifier(%v) = %v wanted %v.", i, tt.in, out, tt.out)
		}
	}
}
