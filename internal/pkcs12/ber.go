// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package pkcs12

import (
	"errors"
	"fmt"
)

// PKCS#12 allows BER, and many certificate authorities ship .p12 files exported with Java
// tools that use it: indefinite lengths, and OCTET STRINGs split in 1000-byte chunks inside a
// constructed OCTET STRING. encoding/asn1 only reads DER, so every structure is passed through
// berToDER before it is parsed (see unmarshal). The MAC is still verified over the original
// bytes: joining chunks does not change the octets of an OCTET STRING.

// maxBERDepth bounds the nesting of the input, which comes from user uploads.
const maxBERDepth = 64

var errBER = errors.New("pkcs12: malformed BER encoding")

// berToDER re-encodes BER as DER: indefinite lengths become definite and constructed universal
// strings become primitive. DER input is returned unchanged, byte for byte.
func berToDER(in []byte) ([]byte, error) {
	var out []byte
	for len(in) > 0 {
		el, n, err := parseBER(in, 0)
		if err != nil {
			return nil, err
		}
		out = append(out, el.encode()...)
		in = in[n:]
	}
	return out, nil
}

// element is a decoded TLV with its content already in DER.
type element struct {
	identifier []byte // tag octets, with the constructed bit as it will be encoded
	content    []byte
}

func (e element) encode() []byte {
	out := append([]byte{}, e.identifier...)
	out = appendLength(out, len(e.content))
	return append(out, e.content...)
}

// parseBER decodes one element at the start of b and returns it and the number of bytes used.
func parseBER(b []byte, depth int) (element, int, error) {
	if depth > maxBERDepth {
		return element{}, 0, fmt.Errorf("%w: nesting deeper than %d", errBER, maxBERDepth)
	}
	if len(b) < 2 {
		return element{}, 0, errBER
	}

	// Identifier octets: high tag numbers continue while the top bit is set.
	idLen := 1
	if b[0]&0x1f == 0x1f {
		for {
			if idLen >= len(b) || idLen > 5 {
				return element{}, 0, errBER
			}
			idLen++
			if b[idLen-1]&0x80 == 0 {
				break
			}
		}
	}
	id := append([]byte{}, b[:idLen]...)
	constructed := id[0]&0x20 != 0
	pos := idLen
	if pos >= len(b) {
		return element{}, 0, errBER
	}

	// Length octets.
	indefinite := false
	length := 0
	switch l := b[pos]; {
	case l == 0x80:
		if !constructed {
			return element{}, 0, fmt.Errorf("%w: indefinite length on a primitive element", errBER)
		}
		indefinite = true
		pos++
	case l < 0x80:
		length = int(l)
		pos++
	default:
		n := int(l & 0x7f)
		if n > 4 || pos+1+n > len(b) {
			return element{}, 0, fmt.Errorf("%w: invalid length", errBER)
		}
		for _, c := range b[pos+1 : pos+1+n] {
			length = length<<8 | int(c)
		}
		pos += 1 + n
	}
	if !indefinite && (length < 0 || pos+length > len(b)) {
		return element{}, 0, fmt.Errorf("%w: length exceeds the data", errBER)
	}

	if !constructed {
		return element{identifier: id, content: b[pos : pos+length]}, pos + length, nil
	}

	// Constructed: convert every child. A definite-length body ends at pos+length; an
	// indefinite one ends with the end-of-contents octets 00 00.
	var children []element
	end := pos + length
	for {
		if indefinite {
			if pos+1 >= len(b) {
				return element{}, 0, fmt.Errorf("%w: missing end-of-contents", errBER)
			}
			if b[pos] == 0 && b[pos+1] == 0 {
				pos += 2
				break
			}
		} else if pos >= end {
			break
		}
		limit := len(b)
		if !indefinite {
			limit = end
		}
		child, n, err := parseBER(b[pos:limit], depth+1)
		if err != nil {
			return element{}, 0, err
		}
		children = append(children, child)
		pos += n
	}

	if isUniversalString(id) {
		// DER encodes strings as primitive: join the chunks of a constructed string.
		var content []byte
		for _, c := range children {
			if c.identifier[0]&0x1f != id[0]&0x1f || c.identifier[0]&0xc0 != 0 {
				return element{}, 0, fmt.Errorf("%w: mixed chunk types in a constructed string", errBER)
			}
			content = append(content, c.content...)
		}
		return element{identifier: []byte{id[0] &^ 0x20}, content: content}, pos, nil
	}

	var content []byte
	for _, c := range children {
		content = append(content, c.encode()...)
	}
	return element{identifier: id, content: content}, pos, nil
}

// isUniversalString reports whether id is a constructed universal string type (OCTET STRING
// or a character string) that DER requires to be primitive. BIT STRING is left alone: joining
// its chunks needs special care with the unused-bits octet and PKCS#12 does not split it.
func isUniversalString(id []byte) bool {
	if len(id) != 1 || id[0]&0xc0 != 0 || id[0]&0x20 == 0 {
		return false
	}
	switch id[0] & 0x1f {
	case 4, // OCTET STRING
		12, // UTF8String
		19, // PrintableString
		20, // T61String
		22, // IA5String
		26, // VisibleString
		30: // BMPString
		return true
	}
	return false
}

// appendLength appends a DER length.
func appendLength(out []byte, n int) []byte {
	if n < 0x80 {
		return append(out, byte(n))
	}
	var digits []byte
	for v := n; v > 0; v >>= 8 {
		digits = append([]byte{byte(v)}, digits...)
	}
	out = append(out, 0x80|byte(len(digits)))
	return append(out, digits...)
}
