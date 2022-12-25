package evm

import (
	"fmt"
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	memory := NewMemory()
	memory.StoreByte(100, 42)

	// first 100 should be initialized to 0
	for i := 0; i < 100; i++ {
		assert.Equal(t, byte(0), memory.data[i])
	}
	assert.Equal(t, memory.data[100], byte(42))

}

func TestMemoryIncrementsForStoreByte(t *testing.T) {
	memory := NewMemory()
	tests := []struct {
		offset   uint64
		expected uint64
	}{
		{
			offset:   0,
			expected: 1,
		},
		{
			offset:   1,
			expected: 1,
		},
		{
			offset:   31,
			expected: 1,
		},
		{
			offset:   32,
			expected: 2,
		},
		{
			offset:   100,
			expected: 4,
		},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("store byte offset %d", tt.offset)
		t.Run(name, func(t *testing.T) {
			memory.StoreByte(tt.offset, 0)
			assert.Equal(t, tt.expected, memory.ActiveWords())
		})
	}

}

func TestMemoryIncrementsForStoreWord(t *testing.T) {
	memory := NewMemory()
	tests := []struct {
		offset   uint64
		expected uint64
	}{
		{
			offset:   0,
			expected: 1,
		},
		{
			offset:   32,
			expected: 2,
		},
		{
			offset:   64,
			expected: 3,
		},
		{
			offset:   65,
			expected: 4, // memory expands to next word
		},
		{
			offset:   95,
			expected: 4,
		},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("store word offset %d", tt.offset)
		t.Run(name, func(t *testing.T) {
			memory.StoreWord(tt.offset, *uint256.NewInt(0))
			assert.Equal(t, tt.expected, memory.ActiveWords())
		})
	}

}

func TestMemoryIncrementsForLoadRange(t *testing.T) {
	memory := NewMemory()
	tests := []struct {
		offset   uint64
		expected uint64
	}{
		{
			offset:   0,
			expected: 1,
		},
		{
			offset:   1,
			expected: 1,
		},
		{
			offset:   31,
			expected: 1,
		},
		{
			offset:   32,
			expected: 2,
		},
		{
			offset:   100,
			expected: 4,
		},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("load range offset %d length 1", tt.offset)
		t.Run(name, func(t *testing.T) {
			memory.LoadRange(tt.offset, 1)
			assert.Equal(t, tt.expected, memory.ActiveWords())
		})
	}
}

func TestMemoryIncrementsForLoadWord(t *testing.T) {
	memory := NewMemory()
	tests := []struct {
		offset   uint64
		expected uint64
	}{
		{
			offset:   0,
			expected: 1,
		},
		{
			offset:   1,
			expected: 2,
		},
		{
			offset:   68,
			expected: 4,
		},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("load range offset %d", tt.offset)
		t.Run(name, func(t *testing.T) {
			memory.LoadWord(tt.offset)
			assert.Equal(t, tt.expected, memory.ActiveWords())
		})
	}
}

func TestMemoryString(t *testing.T) {
	memory := NewMemory()
	memory.StoreByte(0, 0x01)
	memory.StoreByte(1, 0x02)
	memory.StoreByte(2, 0x03)
	memory.StoreByte(3, 0x04)

	assert.Equal(t, fmt.Sprintf("memory: %s", memory.data), memory.String())
}
