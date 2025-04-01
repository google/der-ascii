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
	"strings"
	"testing"
)

var scannerTests = []struct {
	in     string
	tokens []token
	ok     bool
}{
	{
		`# First, the basic kinds of tokens.
SEQUENCE [SEQUENCE] 1 -1 1.2.3.4 .1.2.3.4 ` + "`aabbcc`" + ` "hello" TRUE FALSE { }

# Tokens can be bunched up together.
SEQUENCE[0]{}SEQUENCE}1}-1}1.2}#comment

# Each of these is legal whitespace.
` + "\t\r\n " + `

# Escape sequences.
"\"\n\x42\\"

# Uppercase hex is fine too.
` + "`AABBCC`" + `

# Length modifiers
indefinite long-form:2 adjust-length:10 adjust-length:-10`,
		[]token{
			{Kind: tokenBytes, Value: []byte{0x30}},
			{Kind: tokenBytes, Value: []byte{0x30}},
			{Kind: tokenBytes, Value: []byte{0x01}},
			{Kind: tokenBytes, Value: []byte{0xff}},
			{Kind: tokenBytes, Value: []byte{42, 3, 4}},
			{Kind: tokenBytes, Value: []byte{1, 2, 3, 4}},
			{Kind: tokenBytes, Value: []byte{0xaa, 0xbb, 0xcc}},
			{Kind: tokenBytes, Value: []byte("hello")},
			{Kind: tokenBytes, Value: []byte{0xff}},
			{Kind: tokenBytes, Value: []byte{0x00}},
			{Kind: tokenLeftCurly},
			{Kind: tokenRightCurly},
			{Kind: tokenBytes, Value: []byte{0x30}},
			{Kind: tokenBytes, Value: []byte{0xa0}},
			{Kind: tokenLeftCurly},
			{Kind: tokenRightCurly},
			{Kind: tokenBytes, Value: []byte{0x30}},
			{Kind: tokenRightCurly},
			{Kind: tokenBytes, Value: []byte{0x01}},
			{Kind: tokenRightCurly},
			{Kind: tokenBytes, Value: []byte{0xff}},
			{Kind: tokenRightCurly},
			{Kind: tokenBytes, Value: []byte{42}},
			{Kind: tokenRightCurly},
			{Kind: tokenBytes, Value: []byte{'"', '\n', 0x42, '\\'}},
			{Kind: tokenBytes, Value: []byte{0xaa, 0xbb, 0xcc}},
			{Kind: tokenIndefinite},
			{Kind: tokenLongForm, Length: 2},
			{Kind: tokenAdjustLength, Length: 10},
			{Kind: tokenAdjustLength, Length: -10},
			{Kind: tokenEOF},
		},
		true,
	},
	// Garbage tokens.
	{"SEQUENC", nil, false},
	{"1...2", nil, false},
	{"true", nil, false},
	{"false", nil, false},
	// Unmatched [.
	{"[SEQUENCE", nil, false},
	// Unmatched ".
	{`"`, nil, false},
	// Unmatched `.
	{"`", nil, false},
	// Integer overflow.
	{"999999999999999999999999999999999999999999999999999999999999999", nil, false},
	// Invalid OID.
	{"1.99.1", nil, false},
	// OID component overflow.
	{"1.1.99999999999999999999999999999999999999999999999999999999999999999", nil, false},
	// Bad tag string.
	{"[THIS IS NOT A VALID TAG]", nil, false},
	{"[]", nil, false},
	// Tags may have long-form overrides.
	{"[long-form:2 SEQUENCE]", []token{{Kind: tokenBytes, Value: []byte{0x3f, 0x80, 0x10}}, {Kind: tokenEOF}}, true},
	// Bad hex bytes.
	{"`hi there!`", nil, false},
	// Bad bit characters.
	{"b`hi there!`", nil, false},
	// UTF-16 literals are parsed correctly.
	{
		`u""`,
		[]token{
			{Kind: tokenBytes, Value: []byte{}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		`u"a☃𝄞"`,
		[]token{
			{Kind: tokenBytes, Value: []byte{0x00, 0x61, 0x26, 0x03, 0xd8, 0x34, 0xdd, 0x1e}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		// The same as above, but written with escape characters.
		`u"\x61\u2603\U0001d11e"`,
		[]token{
			{Kind: tokenBytes, Value: []byte{0x00, 0x61, 0x26, 0x03, 0xd8, 0x34, 0xdd, 0x1e}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		// The same as above, but the large character is written as unpaired surrogates.
		`u"\x61\u2603\ud834\udd1e"`,
		[]token{
			{Kind: tokenBytes, Value: []byte{0x00, 0x61, 0x26, 0x03, 0xd8, 0x34, 0xdd, 0x1e}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		`u"\n\"\\"`,
		[]token{
			{Kind: tokenBytes, Value: []byte{0x00, 0x0a, 0x00, 0x22, 0x00, 0x5c}},
			{Kind: tokenEOF},
		},
		true,
	},
	// UTF-32 literals are parsed correctly.
	{
		`U""`,
		[]token{
			{Kind: tokenBytes, Value: []byte{}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		`U"a☃𝄞"`,
		[]token{
			{Kind: tokenBytes, Value: []byte{0x00, 0x00, 0x00, 0x61, 0x00, 0x00, 0x26, 0x03, 0x00, 0x01, 0xd1, 0x1e}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		// The same as above, but written with escape characters.
		`U"\x61\u2603\U0001d11e"`,
		[]token{
			{Kind: tokenBytes, Value: []byte{0x00, 0x00, 0x00, 0x61, 0x00, 0x00, 0x26, 0x03, 0x00, 0x01, 0xd1, 0x1e}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		// UTF-32 literals happily emit unpaired surrogates if you ask them to.
		`U"\ud834\udd1e"`,
		[]token{
			{Kind: tokenBytes, Value: []byte{0x00, 0x00, 0xd8, 0x34, 0x00, 0x00, 0xdd, 0x1e}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		`U"\n\"\\"`,
		[]token{
			{Kind: tokenBytes, Value: []byte{0x00, 0x00, 0x00, 0x0a, 0x00, 0x00, 0x00, 0x22, 0x00, 0x00, 0x00, 0x5c}},
			{Kind: tokenEOF},
		},
		true,
	},
	// Invalid UTF-8 is illegal in a UTF-16 or UTF-32 literal.
	{"u\"\xff\xff\xff\xff\"", nil, false},
	{"U\"\xff\xff\xff\xff\"", nil, false},
	// A correctly-encoded replacement character is fine, however.
	{
		"u\"\xef\xbf\xbd\"",
		[]token{
			{Kind: tokenBytes, Value: []byte{0xff, 0xfd}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		"U\"\xef\xbf\xbd\"",
		[]token{
			{Kind: tokenBytes, Value: []byte{0x00, 0x00, 0xff, 0xfd}},
			{Kind: tokenEOF},
		},
		true,
	},
	// BIT STRING literals are parsed correctly.
	{
		"b``",
		[]token{
			{Kind: tokenBytes, Value: []byte{0x00}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		"b`1`",
		[]token{
			{Kind: tokenBytes, Value: []byte{0x07, 0x100 - (1 << 7)}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		"b`11`",
		[]token{
			{Kind: tokenBytes, Value: []byte{0x06, 0x100 - (1 << 6)}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		"b`111`",
		[]token{
			{Kind: tokenBytes, Value: []byte{0x05, 0x100 - (1 << 5)}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		"b`1111`",
		[]token{
			{Kind: tokenBytes, Value: []byte{0x04, 0x100 - (1 << 4)}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		"b`11111`",
		[]token{
			{Kind: tokenBytes, Value: []byte{0x03, 0x100 - (1 << 3)}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		"b`111111`",
		[]token{
			{Kind: tokenBytes, Value: []byte{0x02, 0x100 - (1 << 2)}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		"b`1111111`",
		[]token{
			{Kind: tokenBytes, Value: []byte{0x01, 0x100 - (1 << 1)}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		"b`1010101001010101`",
		[]token{
			{Kind: tokenBytes, Value: []byte{0x00, 0xaa, 0x55}},
			{Kind: tokenEOF},
		},
		true,
	},
	{
		"b`101010100101`",
		[]token{
			{Kind: tokenBytes, Value: []byte{0x04, 0xaa, 0x50}},
			{Kind: tokenEOF},
		},
		true,
	},
	// We can stick a | in the middle of a BIT STRING to add "explicit" padding.
	{
		"b`101010100|1010101`",
		[]token{
			{Kind: tokenBytes, Value: []byte{0x07, 0xaa, 0x55}},
			{Kind: tokenEOF},
		},
		true,
	},
	// If explicit padding does not end at a byte boundary, the remaining padding
	// bits are zero.
	{
		"b`101010100101|010`",
		[]token{
			{Kind: tokenBytes, Value: []byte{0x04, 0xaa, 0x54}},
			{Kind: tokenEOF},
		},
		true,
	},
	// Padding that passes a byte boundary is an error.
	{"b`0000000|01`", nil, false},
	// Extra |s are an error.
	{"b`0|0|0`", nil, false},
	// Bad or truncated escape sequences.
	{`"\`, nil, false},
	{`"\x`, nil, false},
	{`"\u`, nil, false},
	{`"\U`, nil, false},
	{`"\x1`, nil, false},
	{`"\u123`, nil, false},
	{`"\U1234567`, nil, false},
	{`"\x??"`, nil, false},
	{`"\u????"`, nil, false},
	{`"\U????????"`, nil, false},
	{`"\?"`, nil, false},
	{`u"\`, nil, false},
	{`u"\x`, nil, false},
	{`u"\u`, nil, false},
	{`u"\U`, nil, false},
	{`u"\x1`, nil, false},
	{`u"\u123`, nil, false},
	{`u"\U1234567`, nil, false},
	{`u"\x??"`, nil, false},
	{`u"\u????"`, nil, false},
	{`u"\U????????"`, nil, false},
	{`u"\?"`, nil, false},
	{`U"\`, nil, false},
	{`U"\x`, nil, false},
	{`U"\u`, nil, false},
	{`U"\U`, nil, false},
	{`U"\x1`, nil, false},
	{`U"\u123`, nil, false},
	{`U"\U1234567`, nil, false},
	{`U"\x??"`, nil, false},
	{`U"\u????"`, nil, false},
	{`U"\U????????"`, nil, false},
	{`U"\?"`, nil, false},
	// Long escape sequences are forbidden in byte strings.
	{`"\u1234"`, nil, false},
	{`"\U12345678"`, nil, false},
	// Tokenization works up to a syntax error.
	{`"hello" "world`, []token{{Kind: tokenBytes, Value: []byte("hello")}}, false},
	// Unterminated quotes.
	{`"hello`, nil, false},
	{`u"hello`, nil, false},
	{`U"hello`, nil, false},
	{"b`0101", nil, false},
	// long-form with invalid number.
	{"long-form:", nil, false},
	{"long-form:garbage", nil, false},
	{"long-form:2garbage", nil, false},
	{"long-form:0", nil, false},
	{"long-form:-1", nil, false},
	// adjust-length with invalid number.
	{"adjust-length:", nil, false},
	{"adjust-length:garbage", nil, false},
	{"adjust-length:2garbage", nil, false},
}

func scanAll(in string) (tokens []token, ok bool) {
	scanner := newScanner(in)
	for {
		token, err := scanner.Next()
		if err != nil {
			return
		}
		tokens = append(tokens, token)
		if token.Kind == tokenEOF {
			ok = true
			return
		}
	}
}

func TestScanner(t *testing.T) {
	for i, tt := range scannerTests {
		tokens, ok := scanAll(tt.in)
		if len(tokens) != len(tt.tokens) {
			t.Errorf("%d. output length mismatch. Got %v, wanted %v.", i, len(tokens), len(tt.tokens))
		}

		for j := 0; j < len(tokens) && j < len(tt.tokens); j++ {
			if tokens[j].Kind != tt.tokens[j].Kind {
				t.Errorf("%d. token %d was %s, wanted %s.", i, j, tokens[j].Kind, tt.tokens[j].Kind)
			} else if tokens[j].Kind == tokenBytes && !bytes.Equal(tokens[j].Value, tt.tokens[j].Value) {
				t.Errorf("%d. token %d had value %x, wanted %x.", i, j, tokens[j].Value, tt.tokens[j].Value)
			} else if tokens[j].Kind == tokenLongForm && tokens[j].Length != tt.tokens[j].Length {
				t.Errorf("%d. token %d had length %d, wanted %d.", i, j, tokens[j].Length, tt.tokens[j].Length)
			}
		}

		if ok != tt.ok {
			t.Errorf("%d. success did not match. Got %v, wanted %v.", i, ok, tt.ok)
		}
	}
}

var asciiToDERTests = []struct {
	in  string
	out []byte
	ok  bool
}{
	{"SEQUENCE { INTEGER { 42 } INTEGER { 1 } }", []byte{0x30, 0x06, 0x02, 0x01, 0x2a, 0x02, 0x01, 0x01}, true},
	// Mismatched curlies.
	{"{", nil, false},
	{"}", nil, false},
	// Invalid token.
	{"BOGUS", nil, false},
	// Length modifiers.
	{"[long-form:2 INTEGER] long-form:3 { 42 }", []byte{0x1f, 0x80, 0x02, 0x83, 0x00, 0x00, 0x01, 0x2a}, true},
	{"SEQUENCE indefinite { INTEGER { 42 } }", []byte{0x30, 0x80, 0x02, 0x01, 0x2a, 0x00, 0x00}, true},
	{"SEQUENCE adjust-length:1 {}", []byte{0x30, 0x01}, true},
	{"INTEGER adjust-length:-1 { 0 }", []byte{0x02, 0x00, 0x00}, true},
	{"SEQUENCE long-form:1 adjust-length:1 {}", []byte{0x30, 0x81, 0x01}, true},
	{"SEQUENCE adjust-length:1 long-form:1 {}", []byte{0x30, 0x81, 0x01}, true},
	// Length modifiers that do not modify a length.
	{"indefinite", nil, false},
	{"indefinite SEQUENCE { }", nil, false},
	{"long-form:2", nil, false},
	{"long-form:2 SEQUENCE { }", nil, false},
	{"adjust-length:2", nil, false},
	{"adjust-length:2 SEQUENCE { }", nil, false},
	{"long-form:2 adjust-length:2", nil, false},
	{"long-form:2 adjust-length:2 SEQUENCE { }", nil, false},
	// Conflicting length modifiers.
	{"SEQUENCE adjust-length:1 adjust-length:1 {}", nil, false},
	{"SEQUENCE long-form:1 long-form:2 {}", nil, false},
	{"SEQUENCE long-form:1 adjust-length:2 long-form:3 {}", nil, false},
	{"SEQUENCE long-form:1 indefinite {}", nil, false},
	{"SEQUENCE indefinite indefinite {}", nil, false},
	// Too long of length modifiers.
	{"[long-form:1 99999]", nil, false},
	{"SEQUENCE long-form:1 { `" + strings.Repeat("a", 1024) + "` }", nil, false},
	// Length adjustment overflow and underflow.
	{"OCTET_STRING adjust-length:-1 {}", nil, false},
	{"OCTET_STRING adjust-length:2147483647 { \"a\" }", nil, false},
}

func TestASCIIToDER(t *testing.T) {
	for i, tt := range asciiToDERTests {
		out, err := asciiToDER(tt.in)
		ok := err == nil
		if !tt.ok {
			if ok {
				t.Errorf("%d. asciiToDER(%v) unexpectedly succeeded.", i, tt.in)
			}
		} else {
			if !ok {
				t.Errorf("%d. asciiToDER(%v) unexpectedly failed: %s.", i, tt.in, err)
			} else if !bytes.Equal(out, tt.out) {
				t.Errorf("%d. asciiToDER(%v) = %x wanted %x.", i, tt.in, out, tt.out)
			}
		}
	}
}
