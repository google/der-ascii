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

// package ascii2der implements the DER-ASCII language described in
// https://github.com/google/der-ascii/blob/master/language.txt.
//
// The Scanner type can be used to parse DER-ASCII files and output byte blobs
// that may or may not be valid DER.
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
//
// The zero-value Position represents the first byte of an anonymous input file.
type Position struct {
	Offset int    // Byte offset.
	Line   int    // Line number (zero-indexed).
	Column int    // Column number (zero-indexed byte, not rune, count).
	File   string // Optional file name for pretty-printing.
}

// String converts a Position to a string.
func (p Position) String() string {
	file := p.File
	if file == "" {
		file = "<input>"
	}
	return fmt.Sprintf("%s:%d:%d", file, p.Line+1, p.Column+1)
}

// A tokenKind is a kind of token.
type tokenKind int

const (
	tokenBytes tokenKind = iota
	tokenLeftCurly
	tokenRightCurly
	tokenIndefinite
	tokenLongForm
	tokenComma
	tokenLeftParen
	tokenRightParen
	tokenWord
	tokenEOF
)

// A ParseError may be produced while executing a DER ASCII file, wrapping
// another error along with a position.
//
// Errors produced by functions in this package my by type-asserted to
// ParseError to try and obtain the position at which the error occurred.
type ParseError struct {
	Pos Position
	Err error
}

// Error makes this type into an error type.
func (e *ParseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Err)
}

// Unwrap extracts the inner wrapped error.
//
// See errors.Unwrap().
func (e *ParseError) Unwrap() error {
	return e.Err
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

type Builtin func(args [][]byte) ([]byte, error)

// A Scanner represents parsing state for a DER ASCII file.
//
// A zero-value Scanner is ready to begin parsing (given that Input is set to
// a valid value). However, it is recommended to use NewScanner to create a new
// Scanner, since it can pre-populate fields other than Input with default
// settings.
type Scanner struct {
	// Input is the input text being processed.
	Input string
	// Builtins is a table of builtin functions that can be called with the usual
	// function call syntax in a DER ASCII file. NewScanner will return a Scanner
	// with a pre-populated table consisting of those functions defined in
	// language.txt, but users may add or remove whatever functions they wish.
	Builtins map[string]Builtin
	// Vars is a table of variables that builtins can use to store and retrieve
	// state, such as via the define() and var() builtins.
	Vars map[string][]byte

	// Position is the current position at which parsing should
	// resume. The Offset field is used for indexing into Input; the remaining
	// fields are used for error-reporting.
	pos Position
}

// NewScanner creates a new scanner for parsing the given input.
func NewScanner(input string) *Scanner {
	s := &Scanner{Input: input}
	setDefaultBuiltins(s)
	return s
}

// SetFile sets the file path shown in this Scanner's error reports.
func (s *Scanner) SetFile(path string) {
	s.pos.File = path
}

// Exec consumes tokens until Input is exhausted, returning the resulting
// encoded maybe-DER.
func (s *Scanner) Exec() ([]byte, error) {
	enc, _, err := s.exec(nil)
	return enc, err
}

// isEOF returns whether the cursor is at least n bytes ahead of the end of the
// input.
func (s *Scanner) isEOF(n int) bool {
	return s.pos.Offset+n >= len(s.Input)
}

// advance advances the scanner's cursor n positions.
//
// Unlike just s.pos.Offset += n, this will not proceed beyond the end of the
// string, and will update the line and column information accordingly.
func (s *Scanner) advance(n int) {
	for i := 0; i < n && !s.isEOF(0); i++ {
		if s.Input[s.pos.Offset] == '\n' {
			s.pos.Line++
			s.pos.Column = 0
		} else {
			s.pos.Column++
		}
		s.pos.Offset++
	}
}

// consume advances exactly n times and returns all source bytes between the
// initial cursor position and excluding the final cursor position.
//
// If EOF is reached before all n bytes are consumed, the function returns
// false.
func (s *Scanner) consume(n int) (string, bool) {
	start := s.pos.Offset
	s.advance(n)
	if s.pos.Offset-start != n {
		return "", false
	}

	return s.Input[start:s.pos.Offset], true
}

// consumeUntil advances the cursor until the given byte is seen, returning all
// source bytes between the initial cursor position and excluding the given
// byte. This function will advance past the searched-for byte.
//
// If EOF is reached before the byte is seen, the function returns false.
func (s *Scanner) consumeUntil(b byte) (string, bool) {
	if i := strings.IndexByte(s.Input[s.pos.Offset:], b); i != -1 {
		text, _ := s.consume(i + 1)
		return text[:i], true
	}
	return "", false
}

// parseEscapeSequence parses a DER-ASCII escape sequence, returning the rune
// it escapes.
//
// Valid escapes are:
// \n \" \\ \xNN \uNNNN \UNNNNNNNN
//
// This function assumes that the scanner's cursor is currently on a \ rune.
func (s *Scanner) parseEscapeSequence() (rune, error) {
	s.advance(1) // Skip the \. The caller is assumed to have validated it.
	if s.isEOF(0) {
		return 0, &ParseError{s.pos, errors.New("expected escape character")}
	}

	switch c := s.Input[s.pos.Offset]; c {
	case 'n':
		s.advance(1)
		return '\n', nil
	case '"', '\\':
		s.advance(1)
		return rune(c), nil
	case 'x', 'u', 'U':
		s.advance(1)

		var digits int
		switch c {
		case 'x':
			digits = 2
		case 'u':
			digits = 4
		case 'U':
			digits = 8
		}

		hexes, ok := s.consume(digits)
		if !ok {
			return 0, &ParseError{s.pos, errors.New("unfinished escape sequence")}
		}

		bytes, err := hex.DecodeString(hexes)
		if err != nil {
			return 0, &ParseError{s.pos, err}
		}

		var r rune
		for _, b := range bytes {
			r <<= 8
			r |= rune(b)
		}
		return r, nil
	default:
		return 0, &ParseError{s.pos, fmt.Errorf("unknown escape sequence \\%c", c)}
	}
}

// parseQuotedString parses a UTF-8 string until the next ".
//
// This function assumes that the scanner's cursor is currently on a " rune.
func (s *Scanner) parseQuotedString() (token, error) {
	s.advance(1) // Skip the ". The caller is assumed to have validated it.
	start := s.pos
	var bytes []byte
	for {
		if s.isEOF(0) {
			return token{}, &ParseError{start, errors.New("unmatched \"")}
		}
		switch c := s.Input[s.pos.Offset]; c {
		case '"':
			s.advance(1)
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
			s.advance(1)
			bytes = append(bytes, c)
		}
	}
}

// parseUTF16String parses a UTF-16 string until the next ".
//
// This function assumes that the scanner's cursor is currently on a u followed
// by a " rune.
func (s *Scanner) parseUTF16String() (token, error) {
	s.advance(2) // Skip the u". The caller is assumed to have validated it.
	start := s.pos
	var bytes []byte
	for {
		if s.isEOF(0) {
			return token{}, &ParseError{start, errors.New("unmatched \"")}
		}

		switch s.Input[s.pos.Offset] {
		case '"':
			s.advance(1)
			return token{Kind: tokenBytes, Value: bytes, Pos: start}, nil
		case '\\':
			r, err := s.parseEscapeSequence()
			if err != nil {
				return token{}, err
			}
			bytes = appendUTF16(bytes, r)
		default:
			r, n := utf8.DecodeRuneInString(s.Input[s.pos.Offset:])
			// Note DecodeRuneInString may return utf8.RuneError if there is a
			// legitimate replacement character in the input. The documentation
			// says errors return (RuneError, 0) or (RuneError, 1).
			if r == utf8.RuneError && n <= 1 {
				return token{}, &ParseError{s.pos, errors.New("invalid UTF-8")}
			}
			s.advance(n)
			bytes = appendUTF16(bytes, r)
		}
	}
}

// parseUTF32String parses a UTF-32 string until the next ".
//
// This function assumes that the scanner's cursor is currently on a U followed
// by a " rune.
func (s *Scanner) parseUTF32String() (token, error) {
	s.advance(2) // Skip the U". The caller is assumed to have validated it.
	start := s.pos
	var bytes []byte
	for {
		if s.isEOF(0) {
			return token{}, &ParseError{start, errors.New("unmatched \"")}
		}

		switch s.Input[s.pos.Offset] {
		case '"':
			s.advance(1)
			return token{Kind: tokenBytes, Value: bytes, Pos: start}, nil
		case '\\':
			r, err := s.parseEscapeSequence()
			if err != nil {
				return token{}, err
			}
			bytes = appendUTF32(bytes, r)
		default:
			r, n := utf8.DecodeRuneInString(s.Input[s.pos.Offset:])
			// Note DecodeRuneInString may return utf8.RuneError if there is a
			// legitimate replacement charaacter in the input. The documentation
			// says errors return (RuneError, 0) or (RuneError, 1).
			if r == utf8.RuneError && n <= 1 {
				return token{}, &ParseError{s.pos, errors.New("invalid UTF-8")}
			}
			s.advance(n)
			bytes = appendUTF32(bytes, r)
		}
	}
}

// next lexes the next token.
func (s *Scanner) next() (token, error) {
again:
	if s.isEOF(0) {
		return token{Kind: tokenEOF, Pos: s.pos}, nil
	}

	switch s.Input[s.pos.Offset] {
	case ' ', '\t', '\n', '\r':
		// Skip whitespace.
		s.advance(1)
		goto again
	case '#':
		// Skip to the end of the comment.
		s.advance(1)
		for !s.isEOF(0) {
			wasNewline := s.Input[s.pos.Offset] == '\n'
			s.advance(1)
			if wasNewline {
				break
			}
		}
		goto again
	case '{':
		s.advance(1)
		return token{Kind: tokenLeftCurly, Pos: s.pos}, nil
	case '}':
		s.advance(1)
		return token{Kind: tokenRightCurly, Pos: s.pos}, nil
	case ',':
		s.advance(1)
		return token{Kind: tokenComma, Pos: s.pos}, nil
	case '(':
		s.advance(1)
		return token{Kind: tokenLeftParen, Pos: s.pos}, nil
	case ')':
		s.advance(1)
		return token{Kind: tokenRightParen, Pos: s.pos}, nil
	case '"':
		return s.parseQuotedString()
	case 'u':
		if !s.isEOF(1) && s.Input[s.pos.Offset+1] == '"' {
			return s.parseUTF16String()
		}
	case 'U':
		if !s.isEOF(1) && s.Input[s.pos.Offset+1] == '"' {
			return s.parseUTF32String()
		}
	case 'b':
		if !s.isEOF(1) && s.Input[s.pos.Offset+1] == '`' {
			s.advance(2) // Skip the b`.
			bitStr, ok := s.consumeUntil('`')
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
		s.advance(1)
		hexStr, ok := s.consumeUntil('`')
		if !ok {
			return token{}, &ParseError{s.pos, errors.New("unmatched `")}
		}
		bytes, err := hex.DecodeString(hexStr)
		if err != nil {
			return token{}, &ParseError{s.pos, err}
		}
		return token{Kind: tokenBytes, Value: bytes, Pos: s.pos}, nil
	case '[':
		s.advance(1)
		tagStr, ok := s.consumeUntil(']')
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
	s.advance(1)
loop:
	for !s.isEOF(0) {
		switch s.Input[s.pos.Offset] {
		case ' ', '\t', '\n', '\r', ',', '(', ')', '{', '}', '[', ']', '`', '"', '#':
			break loop
		default:
			s.advance(1)
		}
	}

	symbol := s.Input[start.Offset:s.pos.Offset]

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

	return token{Kind: tokenWord, Value: []byte(symbol), Pos: s.pos}, nil
}

// exec is the main parser loop.
//
// Because we need to consume all of the tokens between delimiters (e.g. for
// computing the length of the contents of {} or counting arguments in ()), this
// function needs to recurse into itself; the left parameter, when non-nil,
// refers to the left delimiter that triggered the recursion.
//
// This function returns when: it sees an EOF; it sees a comma; it sees the
// matching right-delimiter to left. It returns the encoded contents of the the
// recognized tokens and all of the tokens that were recognized, including
// the token that ended parsing.
func (s *Scanner) exec(left *token) ([]byte, []token, error) {
	var out []byte
	var tokens []token
	var lengthModifier *token
	var word *token
	for {
		token, err := s.next()
		if err != nil {
			return nil, nil, err
		}
		tokens = append(tokens, token)
		if lengthModifier != nil && token.Kind != tokenLeftCurly {
			return nil, nil, &ParseError{lengthModifier.Pos, errors.New("length modifier was not followed by '{'")}
		}
		if word != nil && token.Kind != tokenLeftParen {
			return nil, nil, &ParseError{word.Pos, fmt.Errorf("unrecognized symbol %q", string(token.Value))}
		}
		switch token.Kind {
		case tokenBytes:
			out = append(out, token.Value...)
		case tokenLeftCurly:
			child, _, err := s.exec(&token)
			if err != nil {
				return nil, nil, err
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
				return nil, tokens, &ParseError{lengthModifier.Pos, err}
			}
			out = append(out, child...)
			lengthModifier = nil
		case tokenLeftParen:
			if word == nil {
				return nil, tokens, &ParseError{token.Pos, errors.New("missing function name")}
			}
			var args [][]byte
		argLoop:
			for {
				arg, prev, err := s.exec(&token)
				if err != nil {
					return nil, tokens, err
				}
				args = append(args, arg)
				lastToken := prev[len(prev)-1]
				switch lastToken.Kind {
				case tokenComma:
					if len(prev) < 2 {
						return nil, nil, &ParseError{lastToken.Pos, errors.New("function arguments cannot be empty")}
					}
				case tokenRightParen:
					if len(prev) < 2 {
						// Actually foo(), so the argument list is nil.
						args = nil
					}
					break argLoop
				default:
					return nil, nil, &ParseError{lastToken.Pos, errors.New("expected ',' or ')'")}
				}
			}
			bytes, err := s.executeBuiltin(string(word.Value), args)
			if err != nil {
				return nil, nil, err
			}
			word = nil
			out = append(out, bytes...)
		case tokenRightCurly:
			if left != nil && left.Kind == tokenLeftCurly {
				return out, tokens, nil
			}
			return nil, nil, &ParseError{token.Pos, errors.New("unmatched '}'")}
		case tokenRightParen:
			if left != nil && left.Kind == tokenLeftParen {
				return out, tokens, nil
			}
			return nil, nil, &ParseError{token.Pos, errors.New("unmatched '('")}
		case tokenLongForm, tokenIndefinite:
			lengthModifier = &token
		case tokenComma:
			return out, tokens, nil
		case tokenWord:
			word = &token
		case tokenEOF:
			if left == nil {
				return out, tokens, nil
			} else if left.Kind == tokenLeftCurly {
				return nil, nil, &ParseError{left.Pos, errors.New("unmatched '{'")}
			} else {
				return nil, nil, &ParseError{left.Pos, errors.New("unmatched '('")}
			}
		default:
			panic(token)
		}
	}
}

func (s *Scanner) executeBuiltin(name string, args [][]byte) ([]byte, error) {
	builtin, ok := s.Builtins[name]
	if !ok {
		return nil, fmt.Errorf("unrecognized builtin %q", name)
	}

	return builtin(args)
}
