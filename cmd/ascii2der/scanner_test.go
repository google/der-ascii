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
	"testing"
)

func tokenToString(kind tokenKind) string {
	switch kind {
	case tokenBytes:
		return "bytes"
	case tokenLeftCurly:
		return "left-curly"
	case tokenRightCurly:
		return "right-curly"
	case tokenEOF:
		return "EOF"
	default:
		panic(kind)
	}
}

var scannerTests = []struct {
	in     string
	tokens []token
	ok     bool
}{
	{
		`# First, the basic kinds of tokens.
SEQUENCE [SEQUENCE] 1 -1 1.2.3.4 ` + "`aabbcc`" + ` "hello" TRUE FALSE { }

# Tokens can be bunched up together.
SEQUENCE[0]{}SEQUENCE}1}-1}1.2}#comment

# Each of these is legal whitespace.
` + "\t\r\n " + `

# Escape sequences.
"\"\n\x42\\"

# Uppercase hex is fine too.
` + "`AABBCC`",
		[]token{
			{Kind: tokenBytes, Value: []byte{0x30}},
			{Kind: tokenBytes, Value: []byte{0x30}},
			{Kind: tokenBytes, Value: []byte{0x01}},
			{Kind: tokenBytes, Value: []byte{0xff}},
			{Kind: tokenBytes, Value: []byte{42, 3, 4}},
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
	// Bad hex bytes.
	{"`hi there!`", nil, false},
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
				t.Errorf("%d. token %d was %s, wanted %s.", i, j, tokenToString(tokens[j].Kind), tokenToString(tt.tokens[j].Kind))
			} else if tokens[j].Kind == tokenBytes && !bytes.Equal(tokens[j].Value, tt.tokens[j].Value) {
				t.Errorf("%d. token %d had value %x, wanted %x.", i, j, tokens[j].Value, tt.tokens[j].Value)
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
				t.Errorf("%d. asciiToDER(%v) unexpectedly failed.", i, tt.in)
			} else if !bytes.Equal(out, tt.out) {
				t.Errorf("%d. asciiToDER(%v) = %x wanted %x.", i, tt.in, out, tt.out)
			}
		}
	}
}
