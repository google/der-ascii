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
	"encoding/hex"
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf16"

	"github.com/google/der-ascii/lib"
)

// isMadeOfElements returns true if in can be parsed as a series of DER
// elements with no trailing data and false otherwise.
func isMadeOfElements(in []byte) bool {
	var indefiniteCount int
	for len(in) != 0 {
		if indefiniteCount > 0 && len(in) >= 2 && in[0] == 0 && in[1] == 0 {
			in = in[2:]
			indefiniteCount--
			continue
		}

		_, _, indefinite, rest, ok := parseElement(in)
		if !ok {
			return false
		}
		in = rest
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

func bytesToString(in []byte) string {
	if len(in) == 0 {
		return ""
	}

	var asciiCount int
	for _, b := range in {
		if b < 0x80 && (b == '\n' || unicode.IsPrint(rune(b))) {
			asciiCount++
		}
	}

	if float64(asciiCount)/float64(len(in)) > 0.85 {
		return bytesToQuotedString(in)
	}
	return bytesToHexString(in)
}

func bytesToHexString(in []byte) string {
	return fmt.Sprintf("`%s`", hex.EncodeToString(in))
}

func bytesToQuotedString(in []byte) string {
	var out bytes.Buffer
	out.WriteString(`"`)
	for _, b := range in {
		if b == '\n' {
			out.WriteString(`\n`)
		} else if b == '"' {
			out.WriteString(`\"`)
		} else if b == '\\' {
			out.WriteString(`\\`)
		} else if b >= 0x80 || !unicode.IsPrint(rune(b)) {
			fmt.Fprintf(&out, `\x%02x`, b)
		} else {
			out.WriteByte(b)
		}
	}
	out.WriteString(`"`)
	return out.String()
}

func bytesToUTF16String(in []byte) string {
	var out bytes.Buffer
	out.WriteString(`u"`)
	for i := 0; i < len(in)/2; i++ {
		u := rune(in[2*i])<<8 | rune(in[2*i+1])
		if utf16.IsSurrogate(u) && i+1 < len(in)/2 {
			u2 := rune(in[2*i+2])<<8 | rune(in[2*i+3])
			r := utf16.DecodeRune(u, u2)
			if r != unicode.ReplacementChar {
				if unicode.IsPrint(r) {
					out.WriteRune(r)
				} else {
					fmt.Fprintf(&out, `\U%08x`, r)
				}
				i++
				continue
			}
		}

		if u == '\n' {
			out.WriteString(`\n`)
		} else if u == '"' {
			out.WriteString(`\"`)
		} else if u == '\\' {
			out.WriteString(`\\`)
		} else if !utf16.IsSurrogate(u) && unicode.IsPrint(u) {
			out.WriteRune(u)
		} else if u <= 0xff {
			fmt.Fprintf(&out, `\x%02x`, u)
		} else {
			fmt.Fprintf(&out, `\u%04x`, u)
		}
	}
	out.WriteString(`"`)

	// Print the trailing byte if needed.
	if len(in)&1 == 1 {
		fmt.Fprintf(&out, " `\\x%02x`", in[len(in)-1])
	}

	return out.String()
}

func bytesToUTF32String(in []byte) string {
	var out bytes.Buffer
	out.WriteString(`U"`)
	for i := 0; i < len(in)/4; i++ {
		u := rune(in[4*i])<<24 | rune(in[4*i+1])<<16 | rune(in[4*i+2])<<8 | rune(in[4*i+3])
		if u == '\n' {
			out.WriteString(`\n`)
		} else if u == '"' {
			out.WriteString(`\"`)
		} else if u == '\\' {
			out.WriteString(`\\`)
		} else if unicode.IsPrint(u) {
			out.WriteRune(u)
		} else if u <= 0xff {
			fmt.Fprintf(&out, `\x%02x`, u)
		} else if u <= 0xffff {
			fmt.Fprintf(&out, `\u%04x`, u)
		} else {
			fmt.Fprintf(&out, `\U%08x`, u)
		}
	}
	out.WriteString(`"`)

	// Print the trailing bytes if needed.
	if len(in)&3 != 0 {
		fmt.Fprintf(&out, " `")
		for i := len(in) &^ 3; i < len(in); i++ {
			fmt.Fprintf(&out, "\\x%02x", in[i])
		}
		fmt.Fprintf(&out, "`")
	}

	return out.String()
}

func integerToString(in []byte) string {
	v, ok := decodeInteger(in)
	if ok && -100000 <= v && v <= 100000 {
		return strconv.FormatInt(v, 10)
	}
	return bytesToHexString(in)
}

func objectIdentifierToString(in []byte) string {
	oid, ok := decodeObjectIdentifier(in)
	if !ok {
		return bytesToHexString(in)
	}
	var out bytes.Buffer
	for i, v := range oid {
		if i != 0 {
			out.WriteString(".")
		}
		out.WriteString(strconv.FormatUint(uint64(v), 10))
	}
	return out.String()
}

func addLine(out *bytes.Buffer, indent int, value string) {
	for i := 0; i < indent; i++ {
		out.WriteString("  ")
	}
	out.WriteString(value)
	out.WriteString("\n")
}

// derToASCIIImpl disassembles in and writes the result to out with the given
// indent. If stopAtEOC is true, it will stop after an end-of-contents marker
// and return the remaining unprocessed bytes of in.
func derToASCIIImpl(out *bytes.Buffer, in []byte, indent int, stopAtEOC bool) []byte {
	for len(in) != 0 {
		if stopAtEOC && len(in) >= 2 && in[0] == 0 && in[1] == 0 {
			// Emit a `0000` in lieu of a closing base.
			addLine(out, indent-1, "`0000`")
			return in[2:]
		}

		tag, body, indefinite, rest, ok := parseElement(in)
		if !ok {
			// Nothing more to encode. Write the rest as bytes.
			addLine(out, indent, bytesToString(in))
			return nil
		}
		in = rest

		if indefinite {
			// Emit a `80` in lieu of an open brace.
			addLine(out, indent, fmt.Sprintf("%s `80`", tagToString(tag)))
			in = derToASCIIImpl(out, in, indent+1, true)
			continue
		}

		if len(body) == 0 {
			// If the body is empty, skip the newlines.
			addLine(out, indent, fmt.Sprintf("%s {}", tagToString(tag)))
			continue
		}

		if tag.Constructed {
			// If the element is constructed, recurse.
			addLine(out, indent, fmt.Sprintf("%s {", tagToString(tag)))
			derToASCIIImpl(out, body, indent+1, false)
			addLine(out, indent, "}")
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
				addLine(out, indent, fmt.Sprintf("%s { %s }", tagToString(tag), integerToString(body)))
			case "OBJECT_IDENTIFIER":
				if name, ok := objectIdentifierToName(body); ok {
					addLine(out, indent, fmt.Sprintf("# %s", name))
				}
				addLine(out, indent, fmt.Sprintf("%s { %s }", tagToString(tag), objectIdentifierToString(body)))
			case "BOOLEAN":
				var encoded string
				if len(body) == 1 && body[0] == 0x00 {
					encoded = "FALSE"
				} else if len(body) == 1 && body[0] == 0xff {
					encoded = "TRUE"
				} else {
					encoded = bytesToHexString(body)
				}
				addLine(out, indent, fmt.Sprintf("%s { %s }", tagToString(tag), encoded))
			case "BIT_STRING":
				if len(body) > 1 && body[0] == 0 && isMadeOfElements(body[1:]) {
					// X.509 signatures and SPKIs are always logically treated
					// as byte strings, but mistakenly encoded as a BIT STRING.
					// In some cases, these byte strings are DER-encoded
					// structures themselves. Keep parsing if this is detected.
					addLine(out, indent, fmt.Sprintf("%s {", tagToString(tag)))
					// Emit number of unused bits.
					addLine(out, indent+1, "`00`")
					// Emit the remaining as a DER element.
					derToASCIIImpl(out, body[1:], indent+1, false) // Adds a trailing newline.
					addLine(out, indent, "}")
				} else if len(body) > 1 && body[0] < 8 {
					// The first byte is the number of unused bits.
					addLine(out, indent, fmt.Sprintf("%s { %s %s }", tagToString(tag), bytesToString(body[:1]), bytesToString(body[1:])))
				} else {
					addLine(out, indent, fmt.Sprintf("%s { %s }", tagToString(tag), bytesToString(body)))
				}
			case "BMPString":
				addLine(out, indent, fmt.Sprintf("%s { %s }", tagToString(tag), bytesToUTF16String(body)))
			case "UniversalString":
				addLine(out, indent, fmt.Sprintf("%s { %s }", tagToString(tag), bytesToUTF32String(body)))
			default:
				// Keep parsing if the body looks like ASN.1.
				if isMadeOfElements(body) {
					addLine(out, indent, fmt.Sprintf("%s {", tagToString(tag)))
					derToASCIIImpl(out, body, indent+1, false)
					addLine(out, indent, "}")
				} else {
					addLine(out, indent, fmt.Sprintf("%s { %s }", tagToString(tag), bytesToString(body)))
				}
			}
		}
	}
	return nil
}

func derToASCII(in []byte) string {
	var out bytes.Buffer
	derToASCIIImpl(&out, in, 0, false)
	return out.String()
}
