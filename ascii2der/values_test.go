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

var decodeTagStringTests = []struct {
	input string
	tag   lib.Tag
	ok    bool
}{
	{"SEQUENCE", lib.Tag{lib.ClassUniversal, 16, true}, true},
	{"SEQUENCE CONSTRUCTED", lib.Tag{lib.ClassUniversal, 16, true}, true},
	{"SEQUENCE PRIMITIVE", lib.Tag{lib.ClassUniversal, 16, false}, true},
	{"INTEGER", lib.Tag{lib.ClassUniversal, 2, false}, true},
	{"INTEGER CONSTRUCTED", lib.Tag{lib.ClassUniversal, 2, true}, true},
	{"INTEGER PRIMITIVE", lib.Tag{lib.ClassUniversal, 2, false}, true},
	{"2", lib.Tag{lib.ClassContextSpecific, 2, true}, true},
	{"2 PRIMITIVE", lib.Tag{lib.ClassContextSpecific, 2, false}, true},
	{"APPLICATION 2", lib.Tag{lib.ClassApplication, 2, true}, true},
	{"PRIVATE 2", lib.Tag{lib.ClassPrivate, 2, true}, true},
	{"UNIVERSAL 2", lib.Tag{lib.ClassUniversal, 2, true}, true},
	{"UNIVERSAL 2 CONSTRUCTED", lib.Tag{lib.ClassUniversal, 2, true}, true},
	{"UNIVERSAL 2 PRIMITIVE", lib.Tag{lib.ClassUniversal, 2, false}, true},
	{"UNIVERSAL 2 CONSTRUCTED EXTRA", lib.Tag{}, false},
	{"UNIVERSAL 2 EXTRA", lib.Tag{}, false},
	{"UNIVERSAL NOT_A_NUMBER", lib.Tag{}, false},
	{"UNIVERSAL SEQUENCE", lib.Tag{}, false},
	{"UNIVERSAL", lib.Tag{}, false},
	{"SEQUENCE 2", lib.Tag{}, false},
	{"", lib.Tag{}, false},
	{" SEQUENCE", lib.Tag{}, false},
	{"SEQUENCE ", lib.Tag{}, false},
	{"SEQUENCE  CONSTRUCTED", lib.Tag{}, false},
}

func TestDecodeTagString(t *testing.T) {
	for i, tt := range decodeTagStringTests {
		tag, err := decodeTagString(tt.input)
		if tag != tt.tag || (err == nil) != tt.ok {
			t.Errorf("%d. decodeTagString(%v) = %v, err=%s, wanted %v, success=%v", i, tt.input, tag, err, tt.tag, tt.ok)
		}
	}
}
