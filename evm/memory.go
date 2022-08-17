package evm

import (
	"fmt"

	"github.com/holiman/uint256"
)

type Memory struct {
	data []uint8
}

func NewMemory() *Memory {
	m := make([]uint8, 0)
	return &Memory{data: m}
}

func (m *Memory) expandIfNeeded(offset uint64) {
	if offset >= uint64(len(m.data)) {
		for i := 0; i < int(offset-uint64(len(m.data))+1); i++ {
			m.data = append(m.data, 0)
		}
	}
}

func (m *Memory) StoreByte(offset uint64, value uint8) {
	m.expandIfNeeded(offset)
	m.data[offset] = value
}

func (m *Memory) StoreWord(offset uint64, value uint256.Int) {
	m.expandIfNeeded(offset + 31)
	zero := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	copy(m.data[offset:offset+32], zero)
	value.WriteToSlice(m.data[offset : offset+32])
	// similar function in geth
	// https://github.com/ethereum/go-ethereum/blob/a69d4b273d1164637e0edb2cbad2e51325b7e897/core/vm/memory.go#L52-L62

}

func (m *Memory) Store(offset uint64, size uint64, value uint256.Int) {
	// words to write: size / 32 if size is a multiple of 32 -> Store Word
	// store remaining bytes using StoreByte remainder of size/32 times
	value.Bytes()
}

func (m *Memory) LoadByte(offset uint64) uint8 {
	if offset >= uint64(len(m.data)) {
		return 0
	}
	return m.data[offset]
}

func (m *Memory) LoadRange(offset uint64, length uint64) []byte {
	loaded := make([]byte, 0)
	for o := offset; o < offset+length; o++ {
		loaded = append(loaded, m.LoadByte(o))
	}
	return loaded
}

func (m *Memory) LoadWord(offset uint64) *uint256.Int {
	return uint256.NewInt(0).SetBytes(m.LoadRange(offset, 32))
}

func (m *Memory) ActiveWords() uint64 {
	return uint64(len(m.data) / 32)
}

func (m *Memory) String() string {
	return fmt.Sprintf("memory: %s", m.data)
}
