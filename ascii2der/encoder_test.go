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

var appendTagTests = []struct {
	tag     lib.Tag
	encoded []byte
}{
	{lib.Tag{lib.ClassUniversal, 16, true}, []byte{0x30}},
	{lib.Tag{lib.ClassUniversal, 2, false}, []byte{0x02}},
	{lib.Tag{lib.ClassContextSpecific, 1, true}, []byte{0xa1}},
	{lib.Tag{lib.ClassApplication, 1234, true}, []byte{0x7f, 0x89, 0x52}},
}

func TestAppendTag(t *testing.T) {
	for i, tt := range appendTagTests {
		dst := appendTag(nil, tt.tag)
		if !bytes.Equal(dst, tt.encoded) {
			t.Errorf("%d. appendTag(nil, %v) = %v, wanted %v.", i, tt.tag, dst, tt.encoded)
		}

		dst = appendTag(dst, tt.tag)
		if l := len(tt.encoded); len(dst) != l*2 || !bytes.Equal(dst[:l], tt.encoded) || !bytes.Equal(dst[l:], tt.encoded) {
			t.Errorf("%d. appendTag did not preserve existing contents.", i)
		}
	}
}

var appendLengthTests = []struct {
	length  int
	encoded []byte
}{
	{0, []byte{0}},
	{5, []byte{0x05}},
	{0x1f, []byte{0x1f}},
	{0x80, []byte{0x81, 0x80}},
	{0xff, []byte{0x81, 0xff}},
	{0x100, []byte{0x82, 0x01, 0x00}},
	{0xffffff, []byte{0x83, 0xff, 0xff, 0xff}},
}

func TestAppendLength(t *testing.T) {
	for i, tt := range appendLengthTests {
		dst := appendLength(nil, tt.length)
		if !bytes.Equal(dst, tt.encoded) {
			t.Errorf("%d. appendLength(nil, %v) = %v, wanted %v.", i, tt.length, dst, tt.encoded)
		}

		dst = appendLength(dst, tt.length)
		if l := len(tt.encoded); len(dst) != l*2 || !bytes.Equal(dst[:l], tt.encoded) || !bytes.Equal(dst[l:], tt.encoded) {
			t.Errorf("%d. appendLength did not preserve existing contents.", i)
		}
	}
}

var appendIntegerTests = []struct {
	value   int64
	encoded []byte
}{
	{0, []byte{0}},
	{1, []byte{1}},
	{-1, []byte{0xff}},
	{127, []byte{0x7f}},
	{128, []byte{0x00, 0x80}},
	{0x12345678, []byte{0x12, 0x34, 0x56, 0x78}},
	{-127, []byte{0x81}},
	{-128, []byte{0x80}},
	{-129, []byte{0xff, 0x7f}},
}

func TestAppendInteger(t *testing.T) {
	for i, tt := range appendIntegerTests {
		dst := appendInteger(nil, tt.value)
		if !bytes.Equal(dst, tt.encoded) {
			t.Errorf("%d. appendInteger(nil, %v) = %v, wanted %v.", i, tt.value, dst, tt.encoded)
		}

		dst = appendInteger(dst, tt.value)
		if l := len(tt.encoded); len(dst) != l*2 || !bytes.Equal(dst[:l], tt.encoded) || !bytes.Equal(dst[l:], tt.encoded) {
			t.Errorf("%d. appendInteger did not preserve existing contents.", i)
		}
	}
}

var appendObjectIdentifierTests = []struct {
	value   []uint32
	encoded []byte
	ok      bool
}{
	{[]uint32{0, 1}, []byte{1}, true},
	{[]uint32{1, 2, 3, 4, 0, 127, 128, 129}, []byte{42, 3, 4, 0, 0x7f, 0x81, 0x00, 0x81, 0x01}, true},
	{[]uint32{2, 1}, []byte{81}, true},
	{[]uint32{2, math.MaxUint32 - 80}, []byte{0x8f, 0xff, 0xff, 0xff, 0x7f}, true},
	// Invalid OIDs.
	{[]uint32{}, nil, false},
	{[]uint32{1}, nil, false},
	{[]uint32{1, 40}, nil, false},
	{[]uint32{0, 40}, nil, false},
	{[]uint32{3, 1}, nil, false},
	{[]uint32{2, math.MaxUint32 - 79}, nil, false},
}

func TestAppendObjectIdentifier(t *testing.T) {
	for i, tt := range appendObjectIdentifierTests {
		dst, ok := appendObjectIdentifier(nil, tt.value)
		if !tt.ok {
			if ok {
				t.Errorf("%d. appendObjectIdentifier(nil, %v) unexpectedly suceeded.", i, tt.value)
			} else if len(dst) != 0 {
				t.Errorf("%d. appendObjectIdentifier did not preserve input.", i)
			}
		} else if !bytes.Equal(dst, tt.encoded) {
			t.Errorf("%d. appendObjectIdentifier(nil, %v) = %v, wanted %v.", i, tt.value, dst, tt.encoded)
		}

		dst = []byte{0}
		dst, ok = appendObjectIdentifier(dst, tt.value)
		if !tt.ok {
			if ok {
				t.Errorf("%d. appendObjectIdentifier(nil, %v) unexpectedly suceeded.", i, tt.value)
			} else if !bytes.Equal(dst, []byte{0}) {
				t.Errorf("%d. appendObjectIdentifier did not preserve input.", i)
			}
		} else if l := len(tt.encoded); len(dst) != l+1 || dst[0] != 0 || !bytes.Equal(dst[1:], tt.encoded) {
			t.Errorf("%d. appendObjectIdentifier did not preserve existing contents.", i)
		}
	}
}
