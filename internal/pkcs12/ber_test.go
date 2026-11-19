// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package pkcs12

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"math/big"
	"testing"
	"time"
)

func mustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func TestBERToDER(t *testing.T) {
	cases := map[string]struct{ in, want string }{
		"indefinite sequence with a chunked OCTET STRING": {
			// 30 80 { 04 'abc', 24 80 { 04 'de', 04 'f' } 00 00 } 00 00
			in:   "3080" + "0403616263" + "2480" + "04026465" + "040166" + "0000" + "0000",
			want: "300a" + "0403616263" + "0403646566",
		},
		"nested indefinite lengths": {
			in:   "3080" + "3080" + "020101" + "0000" + "0000",
			want: "3005" + "3003" + "020101",
		},
		"context tags are kept, only their length changes": {
			in:   "a080" + "0401ff" + "0000",
			want: "a003" + "0401ff",
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := berToDER(mustHex(c.in))
			if err != nil {
				t.Fatal(err)
			}
			if hex.EncodeToString(got) != c.want {
				t.Fatalf("berToDER = %x, want %s", got, c.want)
			}
		})
	}
}

func TestBERToDERLeavesDERUnchanged(t *testing.T) {
	der, err := asn1.Marshal(struct {
		A int
		B []byte
		C string `asn1:"utf8"`
		D []int
		E asn1.RawValue `asn1:"tag:0,explicit"`
	}{1, bytes.Repeat([]byte{7}, 300), "ñandú", []int{1, 2, 3}, asn1.RawValue{Tag: asn1.TagInteger, Bytes: []byte{5}}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := berToDER(der)
	if err != nil || !bytes.Equal(got, der) {
		t.Fatalf("DER input changed: %x -> %x (%v)", der, got, err)
	}
}

func TestBERToDERRejectsMalformedInput(t *testing.T) {
	deep := bytes.Repeat([]byte{0x30, 0x80}, maxBERDepth+2)
	cases := map[string][]byte{
		"truncated":                      mustHex("3005020101"),
		"missing end-of-contents":        mustHex("3080020101"),
		"indefinite primitive":           mustHex("04800000"),
		"length beyond the data":         mustHex("0484ffffffff00"),
		"too deeply nested":              deep,
		"mixed chunks in a string":       mustHex("2480" + "040161" + "020101" + "0000"),
		"single byte":                    {0x30},
		"high tag number not finished":   mustHex("1f81"),
		"length of length too long":      mustHex("0485" + "0000000001" + "00"),
		"child overruns definite parent": mustHex("3003" + "0405616263"),
	}
	for name, in := range cases {
		if _, err := berToDER(in); !errors.Is(err, errBER) {
			t.Errorf("%s: err = %v, want errBER", name, err)
		}
	}
}

// toJavaBER re-encodes DER the way Java's keytool writes PKCS#12: every constructed element
// with indefinite length, OCTET STRINGs longer than 1000 bytes split in chunks inside a
// constructed OCTET STRING, and a primitive [0] (EncryptedContent) turned into a constructed
// [0] made of OCTET STRING chunks.
func toJavaBER(t *testing.T, der []byte) []byte {
	t.Helper()
	var out []byte
	for len(der) > 0 {
		var raw asn1.RawValue
		rest, err := asn1.Unmarshal(der, &raw)
		if err != nil {
			t.Fatal(err)
		}
		switch {
		case raw.IsCompound:
			out = append(out, raw.FullBytes[0], 0x80)
			out = append(out, toJavaBER(t, raw.Bytes)...)
			out = append(out, 0, 0)
		case raw.Class == asn1.ClassUniversal && raw.Tag == asn1.TagOctetString && len(raw.Bytes) > 1000,
			raw.Class == asn1.ClassContextSpecific && raw.Tag == 0:
			if raw.Class == asn1.ClassContextSpecific {
				out = append(out, 0xa0, 0x80)
			} else {
				out = append(out, 0x24, 0x80)
			}
			for data := raw.Bytes; len(data) > 0; {
				n := min(1000, len(data))
				chunk, _ := asn1.Marshal(data[:n])
				out = append(out, chunk...)
				data = data[n:]
			}
			out = append(out, 0, 0)
		default:
			out = append(out, raw.FullBytes...)
		}
		der = rest
	}
	return out
}

func TestDecodeJavaStyleBERFile(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(7), Subject: pkix.Name{CommonName: "BER COMPANY"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, _ := x509.ParseCertificate(certDER)
	const password = "s3cret"
	// RC2-40 certificates and 3DES keys, as legacy Java exports use.
	derFile, err := LegacyRC2.Encode(key, cert, []*x509.Certificate{cert}, password)
	if err != nil {
		t.Fatal(err)
	}

	// Rebuild it as Java does: the authenticated safe in BER, with the MAC computed over those
	// BER bytes, inside an outer PFX that is BER as well.
	var pfx pfxPdu
	if err := unmarshal(derFile, &pfx); err != nil {
		t.Fatal(err)
	}
	var authSafe []byte
	if err := unmarshal(pfx.AuthSafe.Content.Bytes, &authSafe); err != nil {
		t.Fatal(err)
	}
	berAuthSafe := toJavaBER(t, authSafe)
	if bytes.Equal(berAuthSafe, authSafe) {
		t.Fatal("the authenticated safe was not re-encoded")
	}
	encodedPassword, err := bmpStringZeroTerminated(password)
	if err != nil {
		t.Fatal(err)
	}
	if err := computeMac(&pfx.MacData, berAuthSafe, encodedPassword); err != nil {
		t.Fatal(err)
	}
	octets, err := asn1.Marshal(berAuthSafe)
	if err != nil {
		t.Fatal(err)
	}
	pfx.AuthSafe.Content = asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: octets}
	derPFX, err := asn1.Marshal(pfx)
	if err != nil {
		t.Fatal(err)
	}
	berFile := toJavaBER(t, derPFX)
	if !bytes.HasPrefix(berFile, []byte{0x30, 0x80}) {
		t.Fatalf("the file does not start like a BER export: % x", berFile[:4])
	}

	gotKey, gotCert, caCerts, err := DecodeChain(berFile, password)
	if err != nil {
		t.Fatalf("DecodeChain on a Java-style BER file: %v", err)
	}
	if !gotKey.(*rsa.PrivateKey).Equal(key) || !gotCert.Equal(cert) || len(caCerts) != 1 {
		t.Fatal("the BER file decoded to a different key or certificate")
	}
	if _, _, _, err := DecodeChain(berFile, "wrong"); !errors.Is(err, ErrIncorrectPassword) {
		t.Fatalf("wrong password: err = %v, want ErrIncorrectPassword", err)
	}
}
