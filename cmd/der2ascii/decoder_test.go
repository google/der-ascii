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
	"reflect"
	"testing"

	"github.com/google/der-ascii/internal"
)

var parseTagTests = []struct {
	in  []byte
	tag internal.Tag
	ok  bool
}{
	{[]byte{0x30}, internal.Tag{Number: 16, Constructed: true}, true},
	{[]byte{0x02}, internal.Tag{Number: 2}, true},
	{[]byte{0x7f, 0x89, 0x52}, internal.Tag{Class: internal.ClassApplication, Number: 1234, Constructed: true}, true},
	// Empty.
	{[]byte{}, internal.Tag{}, false},
	// Truncated high-tag-number-form.
	{[]byte{0x7f}, internal.Tag{}, false},
	{[]byte{0x7f, 0xff}, internal.Tag{}, false},
	// Should have been low-tag-number form.
	{[]byte{0x7f, 0x01}, internal.Tag{Class: internal.ClassApplication, Number: 1, Constructed: true, LongFormOverride: 1}, true},
	// Non-minimal encoding.
	{[]byte{0x7f, 0x00}, internal.Tag{Class: internal.ClassApplication, Number: 0, Constructed: true, LongFormOverride: 1}, true},
	{[]byte{0x7f, 0x80, 0x01}, internal.Tag{Class: internal.ClassApplication, Number: 1, Constructed: true, LongFormOverride: 2}, true},
	// Overflow.
	{[]byte{0xff, 0x8f, 0xff, 0xff, 0xff, 0x7f}, internal.Tag{Class: internal.ClassPrivate, Number: (1 << 32) - 1, Constructed: true}, true},
	{[]byte{0xff, 0x9f, 0xff, 0xff, 0xff, 0x7f}, internal.Tag{}, false},
	// Universal tag zero is reserved for EOC, but we parse it here because
	// DER and BER parsers sometimes accept such elements in ANY.
	{[]byte{0x00}, internal.Tag{Number: 0}, true},
	{[]byte{0x20}, internal.Tag{Number: 0, Constructed: true}, true},
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

var sequenceTag = internal.Tag{Class: internal.ClassUniversal, Number: 16, Constructed: true}
var zeroTag = internal.Tag{Class: internal.ClassUniversal, Number: 0, Constructed: false}

var parseTagAndLengthTests = []struct {
	in     []byte
	elem   element
	length int
	ok     bool
}{
	// Short-form length.
	{[]byte{0x30, 0x00}, element{tag: sequenceTag}, 0, true},
	{[]byte{0x30, 0x01}, element{tag: sequenceTag}, 1, true},
	// Indefinite length.
	{[]byte{0x30, 0x80}, element{tag: sequenceTag, indefinite: true}, 0, true},
	// Long-form length.
	{[]byte{0x30, 0x81, 0x80}, element{tag: sequenceTag}, 128, true},
	{[]byte{0x30, 0x81, 0xff}, element{tag: sequenceTag}, 255, true},
	{[]byte{0x30, 0x82, 0x01, 0x00}, element{tag: sequenceTag}, 256, true},
	// Too short.
	{[]byte{0x30}, element{}, 0, false},
	{[]byte{0x30, 0x81}, element{}, 0, false},
	// Non-minimal form length.
	{[]byte{0x30, 0x82, 0x00, 0xff}, element{tag: sequenceTag, longFormOverride: 2}, 255, true},
	{[]byte{0x30, 0x81, 0x1f}, element{tag: sequenceTag, longFormOverride: 1}, 31, true},
	// Overflow.
	{[]byte{0x30, 0x85, 0xff, 0xff, 0xff, 0xff, 0xff}, element{}, 0, false},
	// Empty.
	{[]byte{}, element{}, 0, false},
	// Primitive + indefinite length is illegal.
	{[]byte{0x02, 0x80}, element{}, 0, false},
	// EOC is treated as an element with tag number zero. Although
	// universal tag zero is reserved, DER and BER parsers sometimes accept
	// such elements in ANY.
	{[]byte{0x00, 0x00}, element{tag: zeroTag}, 0, true},
	{[]byte{0x00, 0x01}, element{tag: zeroTag}, 1, true},
}

func TestParseTagAndLength(t *testing.T) {
	for i, tt := range parseTagAndLengthTests {
		elem, length, rest, ok := parseTagAndLength(tt.in)
		if !tt.ok {
			if ok {
				t.Errorf("%d. parseTagAndLength(%v) unexpectedly succeeded.", i, tt.in)
			} else if !bytes.Equal(rest, tt.in) {
				t.Errorf("%d. parseTagAndLength(%v) did not preserve input.", i, tt.in)
			}
		} else {
			if !ok {
				t.Errorf("%d. parseTagAndLength(%v) unexpectedly failed.", i, tt.in)
			} else if !reflect.DeepEqual(elem, tt.elem) || length != tt.length || len(rest) != 0 {
				t.Errorf("%d. parseTagAndLength(%v) = %v, %v, %v wanted %v, %v, [].", i, tt.in, elem, length, rest, tt.elem, tt.length)
			}

			// Test again with trailing data.
			in := make([]byte, len(tt.in)+5)
			copy(in, tt.in)
			elem, length, rest, ok := parseTagAndLength(in)
			if !ok {
				t.Errorf("%d. parseTagAndLength(%v) unexpectedly failed.", i, in)
			} else if !reflect.DeepEqual(elem, tt.elem) || length != tt.length || !bytes.Equal(rest, in[len(tt.in):]) {
				t.Errorf("%d. parseTagAndLength(%v) = %v, %v, %v wanted %v, %v, %v.", i, in, elem, length, rest, tt.elem, tt.length, in[len(tt.in):])
			}
		}
	}
}

var parseElementTests = []struct {
	in   []byte
	elem element
	ok   bool
}{
	// Normal element.
	{[]byte{0x30, 0x00}, element{tag: sequenceTag, body: []byte{}}, true},
	{[]byte{0x30, 0x01, 0xaa}, element{tag: sequenceTag, body: []byte{0xaa}}, true},
	// Indefinite length.
	{[]byte{0x30, 0x80}, element{tag: sequenceTag, indefinite: true}, true},
	// Too short.
	{[]byte{0x30}, element{}, false},
	{[]byte{0x30, 0x01}, element{}, false},
	{[]byte{0x30, 0x81}, element{}, false},
}

func TestParseElement(t *testing.T) {
	for i, tt := range parseElementTests {
		elem, rest, ok := parseElement(tt.in)
		if !tt.ok {
			if ok {
				t.Errorf("%d. parseElement(%v) unexpectedly succeeded.", i, tt.in)
			} else if !bytes.Equal(rest, tt.in) {
				t.Errorf("%d. parseElement(%v) did not preserve input.", i, tt.in)
			}
		} else {
			if !ok {
				t.Errorf("%d. parseElement(%v) unexpectedly failed.", i, tt.in)
			} else if !reflect.DeepEqual(elem, tt.elem) || len(rest) != 0 {
				t.Errorf("%d. parseElement(%v) = %v, %v wanted %v, [].", i, tt.in, elem, rest, tt.elem)
			}

			// Test again with trailing data.
			in := make([]byte, len(tt.in)+5)
			copy(in, tt.in)
			elem, rest, ok := parseElement(in)
			if !ok {
				t.Errorf("%d. parseElement(%v) unexpectedly failed.", i, in)
			} else if !reflect.DeepEqual(elem, tt.elem) || !bytes.Equal(rest, in[len(tt.in):]) {
				t.Errorf("%d. parseElement(%v) = %v, %v wanted %v, %v.", i, in, elem, rest, tt.elem, in[len(tt.in):])
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

var decodeRelativeOIDTests = []struct {
	in  []byte
	out []uint32
	ok  bool
}{
	{[]byte{1}, []uint32{1}, true},
	{[]byte{1, 2, 3, 4, 0x7f, 0x81, 0x00, 0x81, 0x01}, []uint32{1, 2, 3, 4, 127, 128, 129}, true},
	{[]byte{0x8f, 0xff, 0xff, 0xff, 0x7f}, []uint32{math.MaxUint32}, true},
	// Empty.
	{[]byte{}, nil, false},
	// Incomplete component.
	{[]byte{0xff}, nil, false},
	// Overflow.
	{[]byte{0x9f, 0xff, 0xff, 0xff, 0x7f}, nil, false},
}

func TestDecodeRelativeOID(t *testing.T) {
	for i, tt := range decodeRelativeOIDTests {
		out, ok := decodeRelativeOID(tt.in)
		if !tt.ok {
			if ok {
				t.Errorf("%d. decodeRelativeOID(%v) unexpectedly succeeded.", i, tt.in)
			}
		} else if !ok {
			t.Errorf("%d. decodeRelativeOID(%v) unexpectedly failed.", i, tt.in)
		} else if !eqUint32s(out, tt.out) {
			t.Errorf("%d. decodeRelativeOID(%v) = %v wanted %v.", i, tt.in, out, tt.out)
		}
	}
}
