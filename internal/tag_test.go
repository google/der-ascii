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

package internal

import "testing"

var tagGetAliasTests = []struct {
	tag               Tag
	name              string
	toggleConstructed bool
	ok                bool
}{
	{Tag{ClassUniversal, 16, true, 0}, "SEQUENCE", false, true},
	{Tag{ClassUniversal, 16, true, 5}, "SEQUENCE", false, true},
	{Tag{ClassUniversal, 16, false, 0}, "SEQUENCE", true, true},
	{Tag{ClassUniversal, 16, false, 5}, "SEQUENCE", true, true},
	{Tag{ClassUniversal, 2, true, 0}, "INTEGER", true, true},
	{Tag{ClassUniversal, 2, false, 0}, "INTEGER", false, true},
	{Tag{ClassApplication, 2, false, 0}, "", false, false},
	{Tag{ClassUniversal, 0, false, 0}, "", false, false},
}

func TestTagGetAlias(t *testing.T) {
	for i, tt := range tagGetAliasTests {
		name, toggleConstructed, ok := tt.tag.GetAlias()
		if !tt.ok {
			if ok {
				t.Errorf("%d. Unexpectedly found alias for %v.", i, tt.tag)
			}
		} else if !ok {
			t.Errorf("%d. Cound not find alias for %v.", i, tt.tag)
		} else if name != tt.name || toggleConstructed != tt.toggleConstructed {
			t.Errorf("%d. tag.GetAlias = %v, %v, wanted %v, %v.", i, name, toggleConstructed, tt.name, tt.toggleConstructed)
		}
	}
}

var tagByNameTests = []struct {
	name string
	tag  Tag
	ok   bool
}{
	{"BOGUS", Tag{}, false},
	{"SEQUENCE", Tag{ClassUniversal, 16, true, 0}, true},
	{"INTEGER", Tag{ClassUniversal, 2, false, 0}, true},
	{"OCTET STRING", Tag{}, false},
	{"OCTET_STRING", Tag{ClassUniversal, 4, false, 0}, true},
}

func TestTagByName(t *testing.T) {
	for i, tt := range tagByNameTests {
		tag, ok := TagByName(tt.name)
		if !tt.ok {
			if ok {
				t.Errorf("%d. Unexpectedly found tag named %v.", i, tt.name)
			}
		} else if !ok {
			t.Errorf("%d. Cound not find tag named %v.", i, tt.name)
		} else if tag != tt.tag {
			t.Errorf("%d. TagByName(%v) = %v, wanted %v.", i, tt.name, tag, tt.tag)
		}
	}
}
