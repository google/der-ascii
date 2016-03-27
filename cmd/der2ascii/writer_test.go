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
	"testing"

	"github.com/google/der-ascii/lib"
)

var isMadeOfElementsTests = []struct {
	in  []byte
	out bool
}{
	{[]byte{}, true},
	{[]byte{0x00, 0x00}, false},
	{[]byte{0x30, 0x00, 0x02, 0x01, 0x01}, true},
	{[]byte{0x30, 0x80, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01, 0x00, 0x00}, true},
	{[]byte{0x30, 0x80, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01}, false},
}

func TestIsMadeOfElements(t *testing.T) {
	for i, tt := range isMadeOfElementsTests {
		if out := isMadeOfElements(tt.in); out != tt.out {
			t.Errorf("%d. isMadeOfElements(%v) = %v, want %v.", i, tt.in, out, tt.out)
		}
	}
}

var tagToStringTests = []struct {
	in  lib.Tag
	out string
}{
	{lib.Tag{lib.ClassUniversal, 16, true}, "SEQUENCE"},
	{lib.Tag{lib.ClassUniversal, 16, false}, "[SEQUENCE PRIMITIVE]"},
	{lib.Tag{lib.ClassUniversal, 2, true}, "[INTEGER CONSTRUCTED]"},
	{lib.Tag{lib.ClassUniversal, 2, false}, "INTEGER"},
	{lib.Tag{lib.ClassUniversal, 1234, true}, "[UNIVERSAL 1234]"},
	{lib.Tag{lib.ClassContextSpecific, 0, true}, "[0]"},
	{lib.Tag{lib.ClassContextSpecific, 0, false}, "[0 PRIMITIVE]"},
	{lib.Tag{lib.ClassApplication, 0, true}, "[APPLICATION 0]"},
	{lib.Tag{lib.ClassApplication, 0, false}, "[APPLICATION 0 PRIMITIVE]"},
	{lib.Tag{lib.ClassPrivate, 0, true}, "[PRIVATE 0]"},
	{lib.Tag{lib.ClassPrivate, 0, false}, "[PRIVATE 0 PRIMITIVE]"},
}

func TestTagToString(t *testing.T) {
	for i, tt := range tagToStringTests {
		if out := tagToString(tt.in); out != tt.out {
			t.Errorf("%d. tagToString(%v) = %v, want %v.", i, tt.in, out, tt.out)
		}
	}

}

type convertFuncTest struct {
	in  []byte
	out string
}

func testConvertFunc(t *testing.T, name string, convertFunc func(in []byte) string, tests []convertFuncTest) {
	for i, tt := range tests {
		if out := convertFunc(tt.in); out != tt.out {
			t.Errorf("%d. %s(%v) = %v, want %v.", i, name, tt.in, out, tt.out)
		}
	}
}

var bytesToStringTests = []convertFuncTest{
	// Empty strings are empty.
	{nil, ""},
	// Mostly-ASCII strings are encoded in ASCII.
	{[]byte("hello\nworld\n\xff\"\\"), `"hello\nworld\n\xff\"\\"`},
	// Otherwise, encoded in hex.
	{[]byte{0x01, 0x02, 0x03, 0x04, 0x05}, "`0102030405`"},
}

func TestBytesToString(t *testing.T) {
	testConvertFunc(t, "bytesToString", bytesToString, bytesToStringTests)
}

var integerToStringTests = []convertFuncTest{
	// Valid and reasonably-sized integers are encoded as integers.
	{[]byte{42}, "42"},
	{[]byte{0xff}, "-1"},
	// Overly large integers are encoded in hex.
	{[]byte{0xff, 0xff, 0xff, 0xff}, "`ffffffff`"},
	{[]byte{0x00, 0xff, 0xff, 0xff, 0xff}, "`00ffffffff`"},
	// Invalid (non-minimal) integers are encoded in hex.
	{[]byte{0x00, 0x00}, "`0000`"},
}

func TestIntegerToString(t *testing.T) {
	testConvertFunc(t, "integerToString", integerToString, integerToStringTests)
}

var objectIdentifierToStringTests = []convertFuncTest{
	// Prefer to encode OIDs as OIDs.
	{[]byte{42, 3, 4, 5}, "1.2.3.4.5"},
	// Invalid OIDs are encoded in hex.
	{[]byte{0x80, 0x00}, "`8000`"},
}

func TestObjectIdentifierToString(t *testing.T) {
	testConvertFunc(t, "objectIdentifierToString", objectIdentifierToString, objectIdentifierToStringTests)
}

var derToASCIITests = []convertFuncTest{
	// Sample input that tests various difficult cases.
	{
		[]byte{0x30, 0x07, 0x67, 0x61, 0x72, 0x62, 0x61, 0x67, 0x65, 0x04, 0x0a, 0xa0, 0x80, 0x02, 0x01, 0x01, 0x02, 0x01, 0xff, 0x00, 0x00, 0x04, 0x0b, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x30, 0x16, 0xa0, 0x80, 0x02, 0x01, 0x01, 0x02, 0x02, 0x00, 0x00, 0x30, 0x80, 0x05, 0x00, 0x06, 0x03, 0x2a, 0x03, 0x04, 0x06, 0x02, 0x80, 0x00, 0x03, 0x03, 0x00, 0x30, 0x00, 0x03, 0x03, 0x00, 0x00, 0x00, 0x03, 0x05, 0x01, 0x30, 0x80, 0x00, 0x00, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x01, 0xff, 0xff, 0xff, 0xff},
		`SEQUENCE {
  "garbage"
}
OCTET_STRING {
  [0] ` + "`80`" + `
    INTEGER { 1 }
    INTEGER { -1 }
  ` + "`0000`" + `
}
OCTET_STRING { "hello world" }
SEQUENCE {
  [0] ` + "`80`" + `
    INTEGER { 1 }
    INTEGER { ` + "`0000`" + ` }
    SEQUENCE ` + "`80`" + `
      NULL {}
      OBJECT_IDENTIFIER { 1.2.3.4 }
      OBJECT_IDENTIFIER { ` + "`8000`" + ` }
}
BIT_STRING {
  ` + "`00`" + `
  SEQUENCE {}
}
BIT_STRING { ` + "`000000`" + ` }
BIT_STRING { ` + "`0130800000`" + ` }
# sha256
OBJECT_IDENTIFIER { 2.16.840.1.101.3.4.2.1 }
` + "`ffffffff`" + `
`,
	},
}

func TestDERToASCII(t *testing.T) {
	testConvertFunc(t, "derToASCII", derToASCII, derToASCIITests)
}
