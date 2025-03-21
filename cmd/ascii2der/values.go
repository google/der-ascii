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

	"github.com/google/der-ascii/internal"
)

const (
	adjustLengthPrefix = "adjust-length:"
	longFormPrefix     = "long-form:"
)

func isAdjustLength(s string) bool {
	return strings.HasPrefix(s, adjustLengthPrefix)
}

func decodeAdjustLength(s string) (int, error) {
	s, ok := strings.CutPrefix(s, adjustLengthPrefix)
	if !ok {
		return 0, errors.New("not an adjust-length token")
	}

	l, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return l, nil
}

func isLongFormOverride(s string) bool {
	return strings.HasPrefix(s, longFormPrefix)
}

func decodeLongFormOverride(s string) (int, error) {
	s, ok := strings.CutPrefix(s, longFormPrefix)
	if !ok {
		return 0, errors.New("not a long-form override")
	}

	l, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if l <= 0 {
		return 0, errors.New("invalid long-form override")
	}
	return l, nil
}

// decodeTagString decodes s as a tag descriptor and returns the decoded tag or
// an error.
func decodeTagString(s string) (internal.Tag, error) {
	ss := strings.Split(s, " ")

	// Tags may begin with a long-form override.
	var longFormOverride int
	if isLongFormOverride(ss[0]) {
		var err error
		longFormOverride, err = decodeLongFormOverride(ss[0])
		if err != nil {
			return internal.Tag{}, err
		}
		ss = ss[1:]
	}

	if len(ss) == 0 {
		return internal.Tag{}, errors.New("expected tag component")
	}

	// Tag aliases may only be in the first component.
	tag, ok := internal.TagByName(ss[0])
	if ok {
		ss = ss[1:]
	} else {
		// Tags default to constructed, context-specific.
		tag.Class = internal.ClassContextSpecific
		tag.Constructed = true

		// Otherwise, the first component is an optional class.
		switch ss[0] {
		case "APPLICATION":
			tag.Class = internal.ClassApplication
			ss = ss[1:]
		case "PRIVATE":
			tag.Class = internal.ClassPrivate
			ss = ss[1:]
		case "UNIVERSAL":
			tag.Class = internal.ClassUniversal
			ss = ss[1:]
		}

		// The next (or first) component must be the tag number.
		if len(ss) == 0 {
			return internal.Tag{}, errors.New("expected tag number")
		}
		n, err := strconv.ParseUint(ss[0], 10, 32)
		if err != nil {
			return internal.Tag{}, err
		}
		tag.Number = uint32(n)
		ss = ss[1:]
	}

	tag.LongFormOverride = longFormOverride

	// The final token, if any, may be CONSTRUCTED or PRIMITIVE.
	if len(ss) > 0 {
		switch ss[0] {
		case "CONSTRUCTED":
			tag.Constructed = true
		case "PRIMITIVE":
			tag.Constructed = false
		default:
			return internal.Tag{}, fmt.Errorf("unexpected tag component %q", ss[0])
		}
		ss = ss[1:]
	}

	if len(ss) != 0 {
		return internal.Tag{}, fmt.Errorf("excess tag component %q", ss[0])
	}

	return tag, nil
}
