// Package datautil provides utilities for randomness, encode/decode, etc.
package datautil

import (
	srand "crypto/rand"
	"encoding/binary"
	"hash/crc32"
	"math/big"
	irand "math/rand"
	"time"
	"unicode"
)

// EncodeBase62 encodes bytes into a base62 encoded string, returning the
// encoded string; allows for toggling letter case of output string if
// consumer's alphabet is different than Golang's big.Int alphabet; this is the
// dumb way of switching between 0-9,a-z,A-Z and 0-9,A-Z,a-z alphabets.
func EncodeBase62(bytes []byte, invertAlphabet bool) string {
	var i big.Int
	i.SetBytes(bytes)

	if invertAlphabet {
		return flipCase(i.Text(62))
	}

	return i.Text(62)
}

// DecodeBase62 decodes a base62 encoded string, returning a byte slice of the
// given length; if the given length is < 1, the length of the returned byte
// is not guaranteed; if the given length is >= 1, the length of the returned
// byte slice will be of the given length; if the data cannot fit in a byte
// of the given length, a zeroed byte slice of the given length and false will
// be returned; if a decoding error is encountered, a nil byte slice and false
// will be returned; for toggling letter case of input string if producer's
// alphabet is different than Golang's big.Int alphabet; this is the dumb way
// of switching between 0-9,a-z,A-Z and 0-9,A-Z,a-z alphabets.
func DecodeBase62(s string, length int, invertAlphabet bool) ([]byte, bool) {
	text := s
	if invertAlphabet {
		text = flipCase(s)
	}

	var i big.Int
	_, ok := i.SetString(text, 62)
	if !ok {
		return nil, ok
	}

	if length < 1 {
		return i.Bytes(), ok
	}

	bytes := make([]byte, length)

	if len(i.Bytes()) > length {
		return bytes, false
	}

	return i.FillBytes(bytes), true
}

// Simple test if rune is ASCII.
func isASCII(r rune) bool {
	return r <= unicode.MaxASCII
}

// Flips case of ASCII characters in a string; if unicode.SimpleFold() folds a
// rune to a non-ASCII character, we keep folding until we get ASCII.
// For base62 use-case:
// Use this post-encoding when targeting 0-9,A-Z,a-z alphabets;
// Use this pre-decoding when targeting 0-9,A-Z,a-z alphabets;
// Golang's big.Int.SetString() uses a 0-9,a-z,A-Z alphabet.
func flipCase(s string) string {
	split := []rune(s)

	for i, c := range s {
		if isASCII(c) {
			split[i] = unicode.SimpleFold(c)
			// may be unnecessary, but comments say rune can be flipped to non-ascii in some cases:
			// https://github.com/golang/go/blob/6178d25fc0b28724b1b5aec2b1b74fc06d9294c7/src/unicode/letter.go#L331
			for !isASCII(split[i]) {
				split[i] = unicode.SimpleFold(split[i])
			}
		}
	}

	return string(split)
}

// GenerateCrc32Uint32 calculates a CRC32 checksum of the given string,
// returning the checksum as uint32.
func GenerateCrc32Uint32(s string) uint32 {
	crcTable := crc32.MakeTable(crc32.IEEE)

	return crc32.Checksum([]byte(s), crcTable)
}

// Crc32ChecksumBytes calculates a CRC32 checksum of the given string,
// returning the bytes of the checksum.
func Crc32ChecksumBytes(s string) []byte {
	crc := GenerateCrc32Uint32(s)
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, crc)

	return buf
}

// generateInsecureRandomBytes returns a byte slice of the given length w/
// insecurely generated random data.
func generateInsecureRandomBytes(length int) []byte {
	r := irand.New(irand.NewSource(time.Now().UnixNano())) //#nosec:G404
	buf := make([]byte, length)
	_, _ = r.Read(buf) //#nosec:G404

	return buf
}

// GenerateSecureRandomBytes returns a byte slice of the given length w/
// cryptographically secure random data; if an error is encountered while
// attempting to generate secure random data (resulting in a partially filled
// buffter), then the rest of the buffer is filled with insecurely generated
// random data.
func GenerateSecureRandomBytes(length int) []byte {
	buf := make([]byte, length)
	n, err := srand.Read(buf)
	if err != nil {
		insecPad := length - n
		insecureBuf := generateInsecureRandomBytes(insecPad)
		for i := 0; insecPad+i < length; i++ {
			buf[insecPad+i] = insecureBuf[i]
		}
	}

	return buf
}

// SelectInsecureRandomInt selects a random value from the given arguments
// using insecure randomness. Returns 0 if no arguments are given.
func SelectInsecureRandomInt(vals ...int) int {
	length := len(vals)
	if length < 1 {
		return 0
	}

	r := irand.New(irand.NewSource(time.Now().UnixNano())) //#nosec:G404

	return vals[r.Intn(length)] //#nosec:G404
}

// SelectInsecureRandomStr selects a random value from the given arguments
// using insecure randomness. Returns empty string if no arguments are given.
func SelectInsecureRandomStr(vals ...string) string {
	length := len(vals)
	if length < 1 {
		return ""
	}

	r := irand.New(irand.NewSource(time.Now().UnixNano())) //#nosec:G404

	return vals[r.Intn(length)] //#nosec:G404
}
