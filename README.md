# DER ASCII

DER ASCII is a small human-editable language to emit DER-like output. Its goal
is to help create test inputs by taking an existing DER or BER structure,
disassembling it into DER ASCII, making adjustments, and assembling back into
DER. This avoids having to manually fix up all the length prefixes. As a bonus,
it acts as a human-readable view for DER structures.

For the language specification, see [language.txt](/language.txt).

This project provides two tools, `ascii2der` and `der2ascii`, to convert DER
ASCII to a byte string and vice versa.

This is not an official Google project.
