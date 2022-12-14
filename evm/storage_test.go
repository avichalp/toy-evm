package evm

import (
	"strings"
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

// TestStorage tests the Get/Put methods on the Storage struct
func TestStorage(t *testing.T) {
	storage := NewStorage()
	storage.Put(uint256.NewInt(0), uint256.NewInt(1))
	assert.Equal(t, uint256.NewInt(1), storage.Get(*uint256.NewInt(0)))

	// missing value
	assert.Equal(t, uint256.NewInt(0), storage.Get(*uint256.NewInt(1)))
}

func TestStorageString(t *testing.T) {
	storage := NewStorage()
	storage.Put(uint256.NewInt(0), uint256.NewInt(1))
	storage.Put(uint256.NewInt(1), uint256.NewInt(2))
	storage.Put(uint256.NewInt(2), uint256.NewInt(3))

	assert.True(t, strings.Contains(storage.String(), "0: 1"))
	assert.True(t, strings.Contains(storage.String(), "1: 2"))
	assert.True(t, strings.Contains(storage.String(), "2: 3"))
}
