package evm

import (
	"github.com/holiman/uint256"
)

type Calldata struct {
	data []byte
}

func (c *Calldata) ReadByte(offset uint64) uint8 {
	if offset >= uint64(len(c.data)) {
		return 0
	}
	return c.data[offset]
}

func (c *Calldata) ReadWord(offset uint64) *uint256.Int {
	calldataBytes := make([]byte, 32)
	for i := offset; i < offset+32; i++ {
		calldataBytes = append(calldataBytes, c.ReadByte(i))
	}
	return uint256.NewInt(0).SetBytes32(calldataBytes)
}
