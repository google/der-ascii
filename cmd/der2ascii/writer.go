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
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/google/der-ascii/internal"
)

// isMadeOfElements returns true if in can be parsed as a series of BER
// elements with no trailing data and false otherwise.
func isMadeOfElements(in []byte) bool {
	var indefiniteCount int
	for len(in) != 0 {
		if startsWithEOC(in) {
			if indefiniteCount > 0 {
				in = in[2:]
				indefiniteCount--
				continue
			} else {
				// parseElement will parse an unexpected EOC an
				// element with tag number zero. This will
				// cause us to recurse into zero OCTET STRINGs,
				// which is confusing. Require a stricter parse
				// in this function.
				return false
			}
		}

		elem, rest, ok := parseElement(in)
		if !ok {
			return false
		}
		in = rest
		if elem.indefinite {
			indefiniteCount++
		}
	}
	return indefiniteCount == 0
}

func classToString(class internal.Class) string {
	switch class {
	case internal.ClassUniversal:
		return "UNIVERSAL"
	case internal.ClassApplication:
		return "APPLICATION"
	case internal.ClassContextSpecific:
		panic("should not be called")
	case internal.ClassPrivate:
		return "PRIVATE"
	default:
		panic(class)
	}
}

func tagToString(tag internal.Tag) string {
	// Write a short name if possible.
	name, includeConstructed, nameOk := tag.GetAlias()
	if nameOk && tag.LongFormOverride == 0 && !includeConstructed {
		return name
	}
	if !nameOk {
		if tag.Class != internal.ClassContextSpecific {
			name = fmt.Sprintf("%s %d", classToString(tag.Class), tag.Number)
		} else {
			name = fmt.Sprintf("%d", tag.Number)
		}
		includeConstructed = !tag.Constructed
	}
	var components []string
	if tag.LongFormOverride != 0 {
		components = append(components, fmt.Sprintf("long-form:%d", tag.LongFormOverride))
	}
	components = append(components, name)
	if includeConstructed {
		if tag.Constructed {
			components = append(components, "CONSTRUCTED")
		} else {
			components = append(components, "PRIMITIVE")
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(components, " "))
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
		} else if utf8.ValidRune(u) && unicode.IsPrint(u) {
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
		fmt.Fprintf(&out, " `%02x`", in[len(in)-1])
	}

	return out.String()
}

func bytesToUTF32String(in []byte) string {
	var out bytes.Buffer
	out.WriteString(`U"`)
	for i := 0; i < len(in)/4; i++ {
		// Note rune is signed, so we use uint32 here.
		u := uint32(in[4*i])<<24 | uint32(in[4*i+1])<<16 | uint32(in[4*i+2])<<8 | uint32(in[4*i+3])
		if u == '\n' {
			out.WriteString(`\n`)
		} else if u == '"' {
			out.WriteString(`\"`)
		} else if u == '\\' {
			out.WriteString(`\\`)
		} else if utf8.ValidRune(rune(u)) && unicode.IsPrint(rune(u)) {
			out.WriteRune(rune(u))
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
			fmt.Fprintf(&out, "%02x", in[i])
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

func objectIdentifierToName(oid []byte) (string, bool) {
	// TODO(davidben): Now that this list is generated, we may as well sort
	// them in the generator and do a binary search here.
	for _, entry := range oidNames {
		if bytes.Equal(entry.oid, oid) {
			return entry.name, true
		}
	}
	return "", false
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

func relativeOIDToString(in []byte) string {
	oid, ok := decodeRelativeOID(in)
	if !ok {
		return bytesToHexString(in)
	}
	var out bytes.Buffer
	for _, v := range oid {
		out.WriteString(".")
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

func startsWithEOC(in []byte) bool {
	return len(in) >= 2 && in[0] == 0 && in[1] == 0
}

// derToASCIIImpl disassembles in and writes the result to out with the given
// indent. If stopAtEOC is true, it will stop after an end-of-contents marker
// and return the remaining unprocessed bytes of in.
func derToASCIIImpl(out *bytes.Buffer, in []byte, indent int, stopAtEOC bool) []byte {
	for len(in) != 0 {
		if stopAtEOC && startsWithEOC(in) {
			// The caller will consume the EOC.
			return in
		}

		elem, rest, ok := parseElement(in)
		if !ok {
			// Nothing more to encode. Write the rest as bytes.
			addLine(out, indent, bytesToString(in))
			return nil
		}
		in = rest

		if elem.indefinite {
			// If the indefinite-length element is properly closed,
			// we write curly braces with an indefinite modifier.
			// Otherwise, we must write a raw `80` literal. Write
			// the body to a buffer so we may decide this later.
			var child bytes.Buffer
			in = derToASCIIImpl(&child, in, indent+1, true)
			if startsWithEOC(in) {
				addLine(out, indent, fmt.Sprintf("%s indefinite {", tagToString(elem.tag)))
				out.Write(child.Bytes())
				addLine(out, indent, "}")
				in = in[2:]
			} else {
				addLine(out, indent, fmt.Sprintf("%s `80`", tagToString(elem.tag)))
				out.Write(child.Bytes())
			}
			continue
		}

		var header string
		if elem.longFormOverride == 0 {
			header = fmt.Sprintf("%s {", tagToString(elem.tag))
		} else {
			header = fmt.Sprintf("%s long-form:%d {", tagToString(elem.tag), elem.longFormOverride)
		}

		if len(elem.body) == 0 {
			// If the body is empty, skip the newlines.
			addLine(out, indent, fmt.Sprintf("%s}", header))
			continue
		}

		if elem.tag.Constructed {
			// If the element is constructed, recurse.
			addLine(out, indent, header)
			derToASCIIImpl(out, elem.body, indent+1, false)
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
			name, _, _ := elem.tag.GetAlias()
			switch name {
			case "INTEGER":
				addLine(out, indent, fmt.Sprintf("%s %s }", header, integerToString(elem.body)))
			case "OBJECT_IDENTIFIER":
				if name, ok := objectIdentifierToName(elem.body); ok {
					addLine(out, indent, fmt.Sprintf("# %s", name))
				}
				addLine(out, indent, fmt.Sprintf("%s %s }", header, objectIdentifierToString(elem.body)))
			case "RELATIVE_OID":
				addLine(out, indent, fmt.Sprintf("%s %s }", header, relativeOIDToString(elem.body)))
			case "BOOLEAN":
				var encoded string
				if len(elem.body) == 1 && elem.body[0] == 0x00 {
					encoded = "FALSE"
				} else if len(elem.body) == 1 && elem.body[0] == 0xff {
					encoded = "TRUE"
				} else {
					encoded = bytesToHexString(elem.body)
				}
				addLine(out, indent, fmt.Sprintf("%s %s }", header, encoded))
			case "BIT_STRING":
				if len(elem.body) > 1 && elem.body[0] == 0 && isMadeOfElements(elem.body[1:]) {
					// X.509 signatures and SPKIs are always logically treated
					// as byte strings, but mistakenly encoded as a BIT STRING.
					// In some cases, these byte strings are DER-encoded
					// structures themselves. Keep parsing if this is detected.
					addLine(out, indent, header)
					// Emit number of unused bits.
					addLine(out, indent+1, "`00`")
					// Emit the remaining as a DER element.
					derToASCIIImpl(out, elem.body[1:], indent+1, false) // Adds a trailing newline.
					addLine(out, indent, "}")
				} else if len(elem.body) == 1 && elem.body[0] == 0 {
					addLine(out, indent, fmt.Sprintf("%s b`` }", header))
				} else if len(elem.body) > 1 && len(elem.body) <= 5 && elem.body[0] < 8 {
					// Convert to a b`` literal when the leading byte is valid and the
					// number of data octets is at most 4; we limit the length for
					// readability.
					bits := new(strings.Builder)

					// The first octet is the number of unused bits.
					significant := 8 - elem.body[0]
					for i, octet := range elem.body[1:] {
						// Last octet gets some special handling.
						isLast := i == len(elem.body)-2
						for j := 0; j < 8; j++ {
							if isLast && int(significant) == j {
								if octet == 0 {
									break
								}
								bits.WriteRune('|')
							}

							if octet&0x80 == 0 {
								bits.WriteRune('0')
							} else {
								bits.WriteRune('1')
							}
							octet <<= 1
						}
					}

					addLine(out, indent, fmt.Sprintf("%s b`%s` }", header, bits))
				} else if len(elem.body) > 1 && elem.body[0] < 8 {
					// The first byte is the number of unused bits.
					addLine(out, indent, fmt.Sprintf("%s %s %s }", header, bytesToString(elem.body[:1]), bytesToString(elem.body[1:])))
				} else {
					addLine(out, indent, fmt.Sprintf("%s %s }", header, bytesToString(elem.body)))
				}
			case "BMPString":
				addLine(out, indent, fmt.Sprintf("%s %s }", header, bytesToUTF16String(elem.body)))
			case "UniversalString":
				addLine(out, indent, fmt.Sprintf("%s %s }", header, bytesToUTF32String(elem.body)))
			default:
				// Keep parsing if the body looks like ASN.1.
				if isMadeOfElements(elem.body) {
					addLine(out, indent, header)
					derToASCIIImpl(out, elem.body, indent+1, false)
					addLine(out, indent, "}")
				} else {
					addLine(out, indent, fmt.Sprintf("%s %s }", header, bytesToString(elem.body)))
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
