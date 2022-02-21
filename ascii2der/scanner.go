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

package ascii2der

import (
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/der-ascii/internal"
)

// A Position describes a location in the input stream.
type Position struct {
	Offset int // offset, starting at 0
	Line   int // line number, starting at 1
	Column int // column number, starting at 1 (byte count)
}

// A tokenKind is a kind of token.
type tokenKind int

const (
	tokenBytes tokenKind = iota
	tokenLeftCurly
	tokenRightCurly
	tokenIndefinite
	tokenLongForm
	tokenEOF
)

// A ParseError is an error during parsing DER ASCII.
type ParseError struct {
	Pos Position
	Err error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Pos.Line, e.Err)
}

// A token is a token in a DER ASCII file.
type token struct {
	// Kind is the kind of the token.
	Kind tokenKind
	// Value, for a tokenBytes token, is the decoded value of the token in
	// bytes.
	Value []byte
	// Pos is the position of the first byte of the token.
	Pos Position
	// Length, for a tokenLongForm token, is the number of bytes to use to
	// encode the length, not including the initial one.
	Length int
}

var (
	regexpInteger = regexp.MustCompile(`^-?[0-9]+$`)
	regexpOID     = regexp.MustCompile(`^[0-9]+(\.[0-9]+)+$`)
)

type Scanner struct {
	text string
	pos  Position
}

func NewScanner(text string) *Scanner {
	return &Scanner{text: text, pos: Position{Line: 1}}
}

func (s *Scanner) parseEscapeSequence() (rune, error) {
	s.advance() // Skip the \. The caller is assumed to have validated it.
	if s.isEOF() {
		return 0, &ParseError{s.pos, errors.New("expected escape character")}
	}
	switch c := s.text[s.pos.Offset]; c {
	case 'n':
		s.advance()
		return '\n', nil
	case '"', '\\':
		s.advance()
		return rune(c), nil
	case 'x':
		s.advance()
		if s.pos.Offset+2 > len(s.text) {
			return 0, &ParseError{s.pos, errors.New("unfinished escape sequence")}
		}
		b, err := hex.DecodeString(s.text[s.pos.Offset : s.pos.Offset+2])
		if err != nil {
			return 0, &ParseError{s.pos, err}
		}
		s.advanceBytes(2)
		return rune(b[0]), nil
	case 'u':
		s.advance()
		if s.pos.Offset+4 > len(s.text) {
			return 0, &ParseError{s.pos, errors.New("unfinished escape sequence")}
		}
		b, err := hex.DecodeString(s.text[s.pos.Offset : s.pos.Offset+4])
		if err != nil {
			return 0, &ParseError{s.pos, err}
		}
		s.advanceBytes(4)
		return rune(b[0])<<8 | rune(b[1]), nil
	case 'U':
		s.advance()
		if s.pos.Offset+8 > len(s.text) {
			return 0, &ParseError{s.pos, errors.New("unfinished escape sequence")}
		}
		b, err := hex.DecodeString(s.text[s.pos.Offset : s.pos.Offset+8])
		if err != nil {
			return 0, &ParseError{s.pos, err}
		}
		s.advanceBytes(8)
		return rune(b[0])<<24 | rune(b[1])<<16 | rune(b[2])<<8 | rune(b[3]), nil
	default:
		return 0, &ParseError{s.pos, fmt.Errorf("unknown escape sequence \\%c", c)}
	}
}

func (s *Scanner) parseQuotedString() (token, error) {
	s.advance() // Skip the ". The caller is assumed to have validated it.
	start := s.pos
	var bytes []byte
	for {
		if s.isEOF() {
			return token{}, &ParseError{start, errors.New("unmatched \"")}
		}
		switch c := s.text[s.pos.Offset]; c {
		case '"':
			s.advance()
			return token{Kind: tokenBytes, Value: bytes, Pos: start}, nil
		case '\\':
			escapeStart := s.pos
			r, err := s.parseEscapeSequence()
			if err != nil {
				return token{}, err
			}
			if r > 0xff {
				// TODO(davidben): Alternatively, should these encode as UTF-8?
				return token{}, &ParseError{escapeStart, errors.New("illegal escape for quoted string")}
			}
			bytes = append(bytes, byte(r))
		default:
			s.advance()
			bytes = append(bytes, c)
		}
	}
}

func (s *Scanner) parseUTF16String() (token, error) {
	s.advance() // Skip the u. The caller is assumed to have validated it.
	s.advance() // Skip the ". The caller is assumed to have validated it.
	start := s.pos
	var bytes []byte
	for {
		if s.isEOF() {
			return token{}, &ParseError{start, errors.New("unmatched \"")}
		}
		switch c := s.text[s.pos.Offset]; c {
		case '"':
			s.advance()
			return token{Kind: tokenBytes, Value: bytes, Pos: start}, nil
		case '\\':
			r, err := s.parseEscapeSequence()
			if err != nil {
				return token{}, err
			}
			bytes = appendUTF16(bytes, r)
		default:
			r, n := utf8.DecodeRuneInString(s.text[s.pos.Offset:])
			// Note DecodeRuneInString may return utf8.RuneError if there is a
			// legitimate replacement charaacter in the input. The documentation
			// says errors return (RuneError, 0) or (RuneError, 1).
			if r == utf8.RuneError && n <= 1 {
				return token{}, &ParseError{s.pos, errors.New("invalid UTF-8")}
			}
			s.advanceBytes(n)
			bytes = appendUTF16(bytes, r)
		}
	}
}

func (s *Scanner) parseUTF32String() (token, error) {
	s.advance() // Skip the U. The caller is assumed to have validated it.
	s.advance() // Skip the ". The caller is assumed to have validated it.
	start := s.pos
	var bytes []byte
	for {
		if s.isEOF() {
			return token{}, &ParseError{start, errors.New("unmatched \"")}
		}
		switch c := s.text[s.pos.Offset]; c {
		case '"':
			s.advance()
			return token{Kind: tokenBytes, Value: bytes, Pos: start}, nil
		case '\\':
			r, err := s.parseEscapeSequence()
			if err != nil {
				return token{}, err
			}
			bytes = appendUTF32(bytes, r)
		default:
			r, n := utf8.DecodeRuneInString(s.text[s.pos.Offset:])
			// Note DecodeRuneInString may return utf8.RuneError if there is a
			// legitimate replacement charaacter in the input. The documentation
			// says errors return (RuneError, 0) or (RuneError, 1).
			if r == utf8.RuneError && n <= 1 {
				return token{}, &ParseError{s.pos, errors.New("invalid UTF-8")}
			}
			s.advanceBytes(n)
			bytes = appendUTF32(bytes, r)
		}
	}
}

func (s *Scanner) next() (token, error) {
again:
	if s.isEOF() {
		return token{Kind: tokenEOF, Pos: s.pos}, nil
	}

	switch s.text[s.pos.Offset] {
	case ' ', '\t', '\n', '\r':
		// Skip whitespace.
		s.advance()
		goto again
	case '#':
		// Skip to the end of the comment.
		s.advance()
		for !s.isEOF() {
			wasNewline := s.text[s.pos.Offset] == '\n'
			s.advance()
			if wasNewline {
				break
			}
		}
		goto again
	case '{':
		s.advance()
		return token{Kind: tokenLeftCurly, Pos: s.pos}, nil
	case '}':
		s.advance()
		return token{Kind: tokenRightCurly, Pos: s.pos}, nil
	case '"':
		return s.parseQuotedString()
	case 'u':
		if s.pos.Offset+1 < len(s.text) && s.text[s.pos.Offset+1] == '"' {
			return s.parseUTF16String()
		}
	case 'U':
		if s.pos.Offset+1 < len(s.text) && s.text[s.pos.Offset+1] == '"' {
			return s.parseUTF32String()
		}
	case 'b':
		if s.pos.Offset+1 < len(s.text) && s.text[s.pos.Offset+1] == '`' {
			s.advance() // Skip the b.
			s.advance() // Skip the `.
			bitStr, ok := s.consumeUpTo('`')
			if !ok {
				return token{}, &ParseError{s.pos, errors.New("unmatched `")}
			}

			// The leading byte is the number of "extra" bits at the end.
			var bitCount int
			var sawPipe bool
			value := []byte{0}
			for i, r := range bitStr {
				switch r {
				case '0', '1':
					if bitCount%8 == 0 {
						value = append(value, 0)
					}
					if r == '1' {
						value[bitCount/8+1] |= 1 << uint(7-bitCount%8)
					}
					bitCount++
				case '|':
					if sawPipe {
						return token{}, &ParseError{s.pos, errors.New("duplicate |")}
					}

					// bitsRemaining is the number of bits remaining in the output that haven't
					// been used yet. There cannot be more than that many bits past the |.
					bitsRemaining := (len(value)-1)*8 - bitCount
					inputRemaining := len(bitStr) - i - 1
					if inputRemaining > bitsRemaining {
						return token{}, &ParseError{s.pos, fmt.Errorf("expected at most %v explicit padding bits; found %v", bitsRemaining, inputRemaining)}
					}

					sawPipe = true
					value[0] = byte(bitsRemaining)
				default:
					return token{}, &ParseError{s.pos, fmt.Errorf("unexpected rune %q", r)}
				}
			}
			if !sawPipe {
				value[0] = byte((len(value)-1)*8 - bitCount)
			}
			return token{Kind: tokenBytes, Value: value, Pos: s.pos}, nil
		}
	case '`':
		s.advance()
		hexStr, ok := s.consumeUpTo('`')
		if !ok {
			return token{}, &ParseError{s.pos, errors.New("unmatched `")}
		}
		bytes, err := hex.DecodeString(hexStr)
		if err != nil {
			return token{}, &ParseError{s.pos, err}
		}
		return token{Kind: tokenBytes, Value: bytes, Pos: s.pos}, nil
	case '[':
		s.advance()
		tagStr, ok := s.consumeUpTo(']')
		if !ok {
			return token{}, &ParseError{s.pos, errors.New("unmatched [")}
		}
		tag, err := decodeTagString(tagStr)
		if err != nil {
			return token{}, &ParseError{s.pos, err}
		}
		value, err := appendTag(nil, tag)
		if err != nil {
			return token{}, &ParseError{s.pos, err}
		}
		return token{Kind: tokenBytes, Value: value, Pos: s.pos}, nil
	}

	// Normal token. Consume up to the next whitespace character, symbol, or
	// EOF.
	start := s.pos
	s.advance()
loop:
	for !s.isEOF() {
		switch s.text[s.pos.Offset] {
		case ' ', '\t', '\n', '\r', '{', '}', '[', ']', '`', '"', '#':
			break loop
		default:
			s.advance()
		}
	}

	symbol := s.text[start.Offset:s.pos.Offset]

	// See if it is a tag.
	tag, ok := internal.TagByName(symbol)
	if ok {
		value, err := appendTag(nil, tag)
		if err != nil {
			// This is impossible; built-in tags always encode.
			return token{}, &ParseError{s.pos, err}
		}
		return token{Kind: tokenBytes, Value: value, Pos: start}, nil
	}

	if regexpInteger.MatchString(symbol) {
		value, err := strconv.ParseInt(symbol, 10, 64)
		if err != nil {
			return token{}, &ParseError{start, err}
		}
		return token{Kind: tokenBytes, Value: appendInteger(nil, value), Pos: s.pos}, nil
	}

	if regexpOID.MatchString(symbol) {
		oidStr := strings.Split(symbol, ".")
		var oid []uint32
		for _, s := range oidStr {
			u, err := strconv.ParseUint(s, 10, 32)
			if err != nil {
				return token{}, &ParseError{start, err}
			}
			oid = append(oid, uint32(u))
		}
		der, err := appendObjectIdentifier(nil, oid)
		if err != nil {
			return token{}, &ParseError{start, err}
		}
		return token{Kind: tokenBytes, Value: der, Pos: s.pos}, nil
	}

	if symbol == "TRUE" {
		return token{Kind: tokenBytes, Value: []byte{0xff}, Pos: s.pos}, nil
	}

	if symbol == "FALSE" {
		return token{Kind: tokenBytes, Value: []byte{0x00}, Pos: s.pos}, nil
	}

	if symbol == "indefinite" {
		return token{Kind: tokenIndefinite}, nil
	}

	if isLongFormOverride(symbol) {
		l, err := decodeLongFormOverride(symbol)
		if err != nil {
			return token{}, &ParseError{start, err}
		}
		return token{Kind: tokenLongForm, Length: l}, nil
	}

	return token{}, fmt.Errorf("unrecognized symbol %q", symbol)
}

func (s *Scanner) isEOF() bool {
	return s.pos.Offset >= len(s.text)
}

func (s *Scanner) advance() {
	if !s.isEOF() {
		if s.text[s.pos.Offset] == '\n' {
			s.pos.Line++
			s.pos.Column = 0
		} else {
			s.pos.Column++
		}
		s.pos.Offset++
	}
}

func (s *Scanner) advanceBytes(n int) {
	for i := 0; i < n; i++ {
		s.advance()
	}
}

func (s *Scanner) consumeUpTo(b byte) (string, bool) {
	start := s.pos.Offset
	for !s.isEOF() {
		if s.text[s.pos.Offset] == b {
			ret := s.text[start:s.pos.Offset]
			s.advance()
			return ret, true
		}
		s.advance()
	}
	return "", false
}

func (s *Scanner) exec(leftCurly *token) ([]byte, error) {
	var out []byte
	var lengthModifier *token
	for {
		token, err := s.next()
		if err != nil {
			return nil, err
		}
		if lengthModifier != nil && token.Kind != tokenLeftCurly {
			return nil, &ParseError{lengthModifier.Pos, errors.New("length modifier was not followed by '{'")}
		}
		switch token.Kind {
		case tokenBytes:
			out = append(out, token.Value...)
		case tokenLeftCurly:
			child, err := s.exec(&token)
			if err != nil {
				return nil, err
			}
			var lengthOverride int
			if lengthModifier != nil {
				if lengthModifier.Kind == tokenIndefinite {
					out = append(out, 0x80)
					out = append(out, child...)
					out = append(out, 0x00, 0x00)
					lengthModifier = nil
					break
				}
				if lengthModifier.Kind == tokenLongForm {
					lengthOverride = lengthModifier.Length
				}
			}
			out, err = appendLength(out, len(child), lengthOverride)
			if err != nil {
				// appendLength may fail if the lengthModifier was incompatible.
				return nil, &ParseError{lengthModifier.Pos, err}
			}
			out = append(out, child...)
			lengthModifier = nil
		case tokenRightCurly:
			if leftCurly != nil {
				return out, nil
			}
			return nil, &ParseError{token.Pos, errors.New("unmatched '}'")}
		case tokenLongForm, tokenIndefinite:
			lengthModifier = &token
		case tokenEOF:
			if leftCurly == nil {
				return out, nil
			}
			return nil, &ParseError{leftCurly.Pos, errors.New("unmatched '{'")}
		default:
			panic(token)
		}
	}
}

func (s *Scanner) Exec() ([]byte, error) {
	return s.exec(nil)
}
