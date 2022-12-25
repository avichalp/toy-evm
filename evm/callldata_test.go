package evm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestReadByte will test ReadByte function bound to CallData struct.
// The test should check that the correct byte is returned for a given index.
func TestReadByte(t *testing.T) {
	data := fmt.Sprintf("%x",
		[]byte{
			1, 2, 3, 4, 5, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
		})
	callData := NewCalldata(data)
	assert.Equal(t, byte(1), callData.ReadByte(0))
	assert.Equal(t, byte(2), callData.ReadByte(1))
	assert.Equal(t, byte(3), callData.ReadByte(2))
	assert.Equal(t, byte(4), callData.ReadByte(3))
	assert.Equal(t, byte(5), callData.ReadByte(4))
}

// TestReadWord tests the ReadWord function bound to CallData struct.
// The test should check that the correct word is returned for a given offset.
func TestReadWord(t *testing.T) {

	data := fmt.Sprintf("%x",
		[]byte{
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 1, 2, 3, 4, 5,
		})
	calldata := NewCalldata(data)

	assert.Equal(
		t,
		data,
		fmt.Sprintf(
			"0000000000000000000000000000000000000000000000000000000%x",
			calldata.ReadWord(0),
		))
}

// TestSize tests the Size method on Calldata struct
func TestSize(t *testing.T) {
	data := fmt.Sprintf("%x",
		[]byte{
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 1, 2, 3, 4, 5,
		})
	calldata := NewCalldata(data)
	assert.Equal(t, uint64(32), calldata.Size())
}
