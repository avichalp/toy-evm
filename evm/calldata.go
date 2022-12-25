package evm

import (
	"fmt"

	"github.com/holiman/uint256"
)

type Calldata struct {
	data []byte
}

func (c *Calldata) ReadByte(index uint64) uint8 {
	if index >= uint64(len(c.data)) {
		return 0
	}
	return c.data[index]
}

func (c *Calldata) ReadWord(offset uint64) *uint256.Int {
	calldataBytes := make([]byte, 0)
	for i := offset; i < offset+32; i++ {
		calldataBytes = append(calldataBytes, c.ReadByte(i))
	}
	return uint256.NewInt(0).SetBytes32(calldataBytes)
}

// Size returns the lenght of data byte array in Calldata
func (c *Calldata) Size() uint64 {
	return uint64(len(c.data))
}

// Returns the new calldata object
func NewCalldata(calldataHex string) *Calldata {
	// length of calldataHex should be a multiple of 64 (32 bytes)
	if len(calldataHex)%64 != 0 {
		errMsg := fmt.Errorf("calldataHex should be a multiple of 32 bytes")
		panic(errMsg)
	}
	return &Calldata{
		data: HexToBytes(calldataHex),
	}
}
