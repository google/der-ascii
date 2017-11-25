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

	"github.com/google/der-ascii/internal"
)

var decodeTagStringTests = []struct {
	input string
	tag   internal.Tag
	ok    bool
}{
	{"SEQUENCE", internal.Tag{internal.ClassUniversal, 16, true, 0}, true},
	{"long-form:5 SEQUENCE", internal.Tag{internal.ClassUniversal, 16, true, 5}, true},
	{"SEQUENCE CONSTRUCTED", internal.Tag{internal.ClassUniversal, 16, true, 0}, true},
	{"SEQUENCE PRIMITIVE", internal.Tag{internal.ClassUniversal, 16, false, 0}, true},
	{"INTEGER", internal.Tag{internal.ClassUniversal, 2, false, 0}, true},
	{"INTEGER CONSTRUCTED", internal.Tag{internal.ClassUniversal, 2, true, 0}, true},
	{"INTEGER PRIMITIVE", internal.Tag{internal.ClassUniversal, 2, false, 0}, true},
	{"long-form:5 2", internal.Tag{internal.ClassContextSpecific, 2, true, 5}, true},
	{"2 PRIMITIVE", internal.Tag{internal.ClassContextSpecific, 2, false, 0}, true},
	{"APPLICATION 2", internal.Tag{internal.ClassApplication, 2, true, 0}, true},
	{"PRIVATE 2", internal.Tag{internal.ClassPrivate, 2, true, 0}, true},
	{"long-form:5 PRIVATE 2", internal.Tag{internal.ClassPrivate, 2, true, 5}, true},
	{"UNIVERSAL 2", internal.Tag{internal.ClassUniversal, 2, true, 0}, true},
	{"UNIVERSAL 2", internal.Tag{internal.ClassUniversal, 2, true, 0}, true},
	{"UNIVERSAL 2 CONSTRUCTED", internal.Tag{internal.ClassUniversal, 2, true, 0}, true},
	{"UNIVERSAL 2 PRIMITIVE", internal.Tag{internal.ClassUniversal, 2, false, 0}, true},
	{"UNIVERSAL 2 CONSTRUCTED EXTRA", internal.Tag{}, false},
	{"UNIVERSAL 2 EXTRA", internal.Tag{}, false},
	{"UNIVERSAL NOT_A_NUMBER", internal.Tag{}, false},
	{"UNIVERSAL SEQUENCE", internal.Tag{}, false},
	{"UNIVERSAL", internal.Tag{}, false},
	{"SEQUENCE 2", internal.Tag{}, false},
	{"", internal.Tag{}, false},
	{" SEQUENCE", internal.Tag{}, false},
	{"SEQUENCE ", internal.Tag{}, false},
	{"SEQUENCE  CONSTRUCTED", internal.Tag{}, false},
	{"long-form:2", internal.Tag{}, false},
	{"long-form:0 SEQUENCE", internal.Tag{}, false},
	{"long-form:-1 SEQUENCE", internal.Tag{}, false},
	{"long-form:garbage SEQUENCE", internal.Tag{}, false},
}

func TestDecodeTagString(t *testing.T) {
	for i, tt := range decodeTagStringTests {
		tag, err := decodeTagString(tt.input)
		if tag != tt.tag || (err == nil) != tt.ok {
			t.Errorf("%d. decodeTagString(%v) = %v, err=%s, wanted %v, success=%v", i, tt.input, tag, err, tt.tag, tt.ok)
		}
	}
}
