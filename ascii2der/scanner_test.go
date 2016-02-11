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
SEQUENCE [SEQUENCE] 1 -1 1.2.3.4 ` + "`aabbcc`" + ` "hello" { }

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
	// Bad or truncated escape sequences.
	{`"\`, nil, false},
	{`"\x`, nil, false},
	{`"\x1`, nil, false},
	{`"\x??"`, nil, false},
	{`"\?"`, nil, false},
	// Tokenization works up to a syntax error.
	{`"hello" "world`, []token{{Kind: tokenBytes, Value: []byte("hello")}}, false},
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
