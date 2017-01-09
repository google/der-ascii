# DER ASCII

[![Build Status](https://travis-ci.org/google/der-ascii.svg?branch=master)](https://travis-ci.org/google/der-ascii)

DER ASCII is a small human-editable language to emit DER
([Distinguished Encoding Rules](https://en.wikipedia.org/wiki/X.690#DER_encoding))
or BER
([Basic Encoding Rules](https://en.wikipedia.org/wiki/X.690#BER_encoding))
encodings of
[ASN.1](https://en.wikipedia.org/wiki/Abstract_Syntax_Notation_One)
structures and malformed variants of them.

It provides two tools, `ascii2der` and `der2ascii`, to convert DER ASCII to a
byte string and vice versa. To install them, run:

    go get github.com/google/der-ascii/cmd/...

These tools may be used to create test inputs by taking an existing DER or BER
structure, disassembling it with `der2ascii` into DER ASCII, making
adjustments, and assembling back into binary with `ascii2der`. This avoids
having to manually fix up all the length prefixes.  As a bonus, it acts as a
human-readable view for DER structures.

For the language specification and basic examples, see
[language.txt](/language.txt). The [samples](/samples) directory includes
more complex examples from real inputs.

This is not an official Google project.
