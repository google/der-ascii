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
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/der-ascii/lib"
)

// decodeTagString decodes s as a tag descriptor and returns the decoded tag or
// an error.
func decodeTagString(s string) (lib.Tag, error) {
	ss := strings.Split(s, " ")

	// Tag aliases may only be in the first component.
	tag, ok := lib.TagByName(ss[0])
	if ok {
		ss = ss[1:]
		goto constructedOrPrimitive
	}

	// Tags default to constructed, context-specific.
	tag.Class = lib.ClassContextSpecific
	tag.Constructed = true

	// Otherwise, the first component is an optional class.
	switch ss[0] {
	case "APPLICATION":
		tag.Class = lib.ClassApplication
		ss = ss[1:]
	case "PRIVATE":
		tag.Class = lib.ClassPrivate
		ss = ss[1:]
	case "UNIVERSAL":
		tag.Class = lib.ClassUniversal
		ss = ss[1:]
	}

	{
		// The next (or first) component must be the tag number.
		// Introduce a scope so the goto above is legal.
		if len(ss) == 0 {
			return lib.Tag{}, errors.New("expected tag number")
		}
		n, err := strconv.ParseUint(ss[0], 10, 32)
		if err != nil {
			return lib.Tag{}, err
		}
		tag.Number = uint32(n)
		ss = ss[1:]
	}

constructedOrPrimitive:
	// The final token, if any, may be CONSTRUCTED or PRIMITIVE.
	if len(ss) > 0 {
		switch ss[0] {
		case "CONSTRUCTED":
			tag.Constructed = true
		case "PRIMITIVE":
			tag.Constructed = false
		default:
			return lib.Tag{}, fmt.Errorf("unexpected tag component '%s'", ss[0])
		}
		ss = ss[1:]
	}

	if len(ss) != 0 {
		return lib.Tag{}, fmt.Errorf("excess tag component '%s'", ss[0])
	}

	return tag, nil
}
