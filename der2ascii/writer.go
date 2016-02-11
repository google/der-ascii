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
	"encoding/hex"
	"fmt"
	"strconv"
	"unicode"

	"github.com/google/der-ascii/lib"
)

type writer struct {
	out    string
	indent int
}

func (w *writer) String() string {
	return w.out
}

func (w *writer) SetIndent(indent int) {
	w.indent = indent
}

func (w *writer) Indent() int {
	return w.indent
}

func (w *writer) AddIndent(v int) {
	w.indent += v
}

func (w *writer) WriteLine(line string) {
	for i := 0; i < w.indent; i++ {
		w.out += "  "
	}
	w.out += line
	w.out += "\n"
}

// isMadeOfElements returns true if bytes can be parsed as a series of DER
// elements with no trailing data and false otherwise.
func isMadeOfElements(bytes []byte) bool {
	var indefiniteCount int
	for len(bytes) != 0 {
		if indefiniteCount > 0 && len(bytes) >= 2 && bytes[0] == 0 && bytes[1] == 0 {
			bytes = bytes[2:]
			indefiniteCount--
			continue
		}

		_, _, indefinite, rest, ok := parseElement(bytes)
		if !ok {
			return false
		}
		bytes = rest
		if indefinite {
			indefiniteCount++
		}
	}
	return indefiniteCount == 0
}

func classToString(class lib.Class) string {
	switch class {
	case lib.ClassUniversal:
		return "UNIVERSAL"
	case lib.ClassApplication:
		return "APPLICATION"
	case lib.ClassContextSpecific:
		panic("should not be called")
	case lib.ClassPrivate:
		return "PRIVATE"
	default:
		panic(class)
	}
}

func tagToString(tag lib.Tag) string {
	// Write a short name if possible.
	name, toggleConstructed, ok := tag.GetAlias()
	if ok {
		if !toggleConstructed {
			return name
		}
		constructed := "PRIMITIVE"
		if tag.Constructed {
			constructed = "CONSTRUCTED"
		}
		return fmt.Sprintf("[%s %s]", name, constructed)
	}

	out := "["
	if tag.Class != lib.ClassContextSpecific {
		out += fmt.Sprintf("%s ", classToString(tag.Class))
	}
	out += fmt.Sprintf("%d", tag.Number)
	if !tag.Constructed {
		out += " PRIMITIVE"
	}
	out += "]"
	return out
}

func bytesToString(bytes []byte) string {
	if len(bytes) == 0 {
		return ""
	}

	var asciiCount int
	for _, b := range bytes {
		if b < 0x80 && (b == '\n' || unicode.IsPrint(rune(b))) {
			asciiCount++
		}
	}

	if float64(asciiCount)/float64(len(bytes)) > 0.85 {
		return bytesToQuotedString(bytes)
	} else {
		return bytesToHexString(bytes)
	}
}

func bytesToHexString(bytes []byte) string {
	return fmt.Sprintf("`%s`", hex.EncodeToString(bytes))
}

func bytesToQuotedString(bytes []byte) string {
	out := `"`
	for _, b := range bytes {
		if b == '\n' {
			out += `\n`
		} else if b == '"' {
			out += `\"`
		} else if b == '\\' {
			out += `\\`
		} else if b >= 0x80 || !unicode.IsPrint(rune(b)) {
			out += fmt.Sprintf(`\x%02x`, b)
		} else {
			out += string([]byte{b})
		}
	}
	out += `"`
	return out
}

func integerToString(bytes []byte) string {
	v, ok := decodeInteger(bytes)
	if ok && -100000 <= v && v <= 100000 {
		return strconv.FormatInt(v, 10)
	}
	return bytesToHexString(bytes)
}

func objectIdentifierToString(bytes []byte) string {
	oid, ok := decodeObjectIdentifier(bytes)
	if !ok {
		return bytesToHexString(bytes)
	}
	var out string
	for i, v := range oid {
		if i != 0 {
			out += "."
		}
		out += strconv.FormatUint(uint64(v), 10)
	}
	return out
}

func derToASCIIImpl(w *writer, bytes []byte, stopAtEOC bool) []byte {
	for len(bytes) != 0 {
		if stopAtEOC && len(bytes) >= 2 && bytes[0] == 0 && bytes[1] == 0 {
			// Emit a `0000` in lieu of a closing base.
			w.AddIndent(-1)
			w.WriteLine(bytesToString(bytes[:2]))
			return bytes[2:]
		}

		tag, body, indefinite, rest, ok := parseElement(bytes)
		if !ok {
			// Nothing more to encode. Write the rest as bytes.
			w.WriteLine(bytesToString(bytes))
			return nil
		}
		bytes = rest

		if indefinite {
			// Emit a `80` in lieu of an open brace.
			w.WriteLine(fmt.Sprintf("%s `80`", tagToString(tag)))
			indent := w.Indent()
			w.AddIndent(1)
			bytes = derToASCIIImpl(w, bytes, true)
			// If EOC was missing, the indent may not have been
			// restored correctly.
			w.SetIndent(indent)
			continue
		}

		if len(body) == 0 {
			// If the body is empty, skip the newlines.
			w.WriteLine(fmt.Sprintf("%s {}", tagToString(tag)))
			continue
		}

		if tag.Constructed {
			// If the element is constructed, recurse.
			w.WriteLine(fmt.Sprintf("%s {", tagToString(tag)))
			w.AddIndent(1)
			derToASCIIImpl(w, body, false)
			w.AddIndent(-1)
			w.WriteLine("}")
		} else {
			// The element is primitive. By default, emit the body
			// on the same line as curly braces. However, in some
			// cases, we heuristically decode the body as DER too.
			// In this case, the newlines are inserted as in the
			// constructed case.

			// If ok is false, name will be empty. There is also no
			// need to check toggleConstructed as we already know
			// the tag is primitive.
			name, _, _ := tag.GetAlias()
			switch name {
			case "INTEGER":
				w.WriteLine(fmt.Sprintf("%s { %s }", tagToString(tag), integerToString(body)))
			case "OBJECT_IDENTIFIER":
				w.WriteLine(fmt.Sprintf("%s { %s }", tagToString(tag), objectIdentifierToString(body)))
			case "BIT_STRING":
				// X.509 encodes signatures and SPKIs in BIT
				// STRINGs, so there is a 0 phase byte followed
				// by the potentially DER-encoded structure.
				if len(body) > 1 && body[0] == 0 && isMadeOfElements(body[1:]) {
					w.WriteLine(fmt.Sprintf("%s {", tagToString(tag)))
					w.AddIndent(1)
					// Emit the phase byte.
					w.WriteLine(bytesToString(body[:1]))
					// Emit the remaining as a DER element.
					derToASCIIImpl(w, body[1:], false) // Adds a trailing newline.
					w.AddIndent(-1)
					w.WriteLine("}")
				} else {
					w.WriteLine(fmt.Sprintf("%s { %s }", tagToString(tag), bytesToString(body)))
				}
			default:
				// Keep parsing if the body looks like ASN.1.
				//
				// TODO(davidben): This is O(N^2) for deeply-
				// nested indefinite-length encodings inside
				// primitive elements.
				if isMadeOfElements(body) {
					w.WriteLine(fmt.Sprintf("%s {", tagToString(tag)))
					w.AddIndent(1)
					derToASCIIImpl(w, body, false)
					w.AddIndent(-1)
					w.WriteLine("}")
				} else {
					w.WriteLine(fmt.Sprintf("%s { %s }", tagToString(tag), bytesToString(body)))
				}
			}
		}
	}
	return nil
}

func derToASCII(bytes []byte) string {
	var w writer
	derToASCIIImpl(&w, bytes, false)
	return w.String()
}
