package evm

import (
	"fmt"
	"math"

	"github.com/holiman/uint256"
)

var zeroWord = []byte{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

type Memory struct {
	data []uint8
}

func NewMemory() *Memory {
	m := make([]uint8, 0)
	return &Memory{data: m}
}

// expandIfNeeded will expand the memory byte slice when
// the vm encounters either an MLOAD or MSTORE/MSTORE8
// instructions
//
// According to the Yellow Paper, the number of active words
// is expanded when we are reading or writing a previously
// untouched memory location
//
// In case of MLOAD:
//
//	μ′i ≡ max(μi,⌈(μs[0]+32)÷32⌉)
//
// In case of MSTORE:
//
//	μ′i ≡ max(μi, ⌈(μs[0]+1)÷32⌉)
//
// μi: highest memory location
// μs: stack (μs[0] is the top of the stack)
func (m *Memory) expandIfNeeded(offset uint64) {
	if offset < uint64(len(m.data)) {
		return // No need to grow memory
	}
	activeWordsBefore := m.ActiveWords()
	activeWordsAfter := uint64(math.Max(float64(activeWordsBefore), math.Ceil(float64(offset+1)/float64(32))))
	for i := 0; i < int(activeWordsAfter-activeWordsBefore); i++ {
		m.data = append(m.data, zeroWord...)
	}
}

func (m *Memory) StoreByte(offset uint64, value uint8) {
	m.expandIfNeeded(offset)
	m.data[offset] = value
}

func (m *Memory) StoreWord(offset uint64, value uint256.Int) {
	m.expandIfNeeded(offset + 31)
	value.WriteToSlice(m.data[offset : offset+32])
}

func (m *Memory) LoadRange(offset uint64, length uint64) []byte {
	m.expandIfNeeded(offset + length - 1)
	return m.data[offset : offset+length]
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
