package main

import (
	"github.com/holiman/uint256"
)

type Calldata struct {
	data []byte
}

func (c Calldata) Read(offset uint64) uint8 {
	// todo: use bufio.Reader
	if offset >= uint64(len(c.data)) {
		return 0
	}
	return c.data[offset]
}

func (c Calldata) ReadWord(offset uint64) *uint256.Int {
	calldataBytes := make([]byte, 32)
	for i := offset; i < offset+32; i++ {
		calldataBytes = append(calldataBytes, c.Read(i))
	}
	return uint256.NewInt(0).SetBytes32(calldataBytes)
}
