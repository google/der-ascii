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

// Package internal contains common routines between der2ascii and ascii2der.
package internal

type Class byte

const (
	ClassUniversal       Class = 0x0
	ClassApplication     Class = 0x40
	ClassContextSpecific Class = 0x80
	ClassPrivate         Class = 0xc0
)

type Tag struct {
	Class       Class
	Number      uint32
	Constructed bool
	// LongFormOverride, if non-zero, is how many bytes this tag is encoded
	// with in long form, excluding the initial byte.
	LongFormOverride int
}

// GetAlias looks up the alias for the given tag. If one exists, it returns the
// name and sets toggleConstructed if the tag's constructed bit does not match
// the alias's default. Otherwise it sets ok to false.
func (t Tag) GetAlias() (name string, toggleConstructed bool, ok bool) {
	if t.Class != ClassUniversal {
		return
	}
	for _, u := range universalTags {
		if u.number == t.Number {
			name = u.name
			toggleConstructed = u.constructed != t.Constructed
			ok = true
			return
		}
	}
	return
}

var universalTags = []struct {
	number      uint32
	name        string
	constructed bool
}{
	// 0 is reserved.
	{1, "BOOLEAN", false},
	{2, "INTEGER", false},
	{3, "BIT_STRING", false},
	{4, "OCTET_STRING", false},
	{5, "NULL", false},
	{6, "OBJECT_IDENTIFIER", false},
	{7, "OBJECT_DESCRIPTOR", false},
	{8, "EXTERNAL", false},
	{9, "REAL", false},
	{10, "ENUMERATED", false},
	{11, "EMBEDDED_PDV", false},
	{12, "UTF8String", false},
	{13, "RELATIVE_OID", false},
	{14, "TIME", false},
	// 15 is reserved for future expansion.
	{16, "SEQUENCE", true},
	{17, "SET", true},
	{18, "NumericString", false},
	{19, "PrintableString", false},
	{20, "T61String", false},
	{21, "VideotexString", false},
	{22, "IA5String", false},
	{23, "UTCTime", false},
	{24, "GeneralizedTime", false},
	{25, "GraphicString", false},
	{26, "VisibleString", false},
	{27, "GeneralString", false},
	{28, "UniversalString", false},
	{30, "BMPString", false},
	{31, "DATE", false},
	{32, "TIME-OF-DAY", false},
	{33, "DATE-TIME", false},
	{34, "DURATION", false},
	{35, "OID-IRI", false},
	{36, "RELATIVE-OID-IRI", false},
}

// TagByName returns the universal tag by name or false if no tag matches.
func TagByName(name string) (Tag, bool) {
	for _, u := range universalTags {
		if u.name == name {
			return Tag{ClassUniversal, u.number, u.constructed, 0}, true
		}
	}
	return Tag{}, false
}
