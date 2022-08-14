package evm

import (
	"encoding/hex"
	"regexp"
)

// HexToBytes convert a hex string to a byte sequence.
// The hex string can have spaces between bytes.
func HexToBytes(s string) []byte {
	s = regexp.MustCompile(" ").ReplaceAllString(s, "")
	b := make([]byte, hex.DecodedLen(len(s)))
	_, err := hex.Decode(b, []byte(s))
	if err != nil {
		panic(err)
	}

	return b[:]
}
