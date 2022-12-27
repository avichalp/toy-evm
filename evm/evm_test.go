package evm

import (
	"fmt"
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

type expected struct {
	stack      []*uint256.Int
	memory     []byte
	returndata []byte
	storage    map[uint256.Int]*uint256.Int
	gasLeft    uint64
}

func TestRunSuccess(t *testing.T) {
	Init()
	t.Cleanup(func() {
		InstructionSet = make(map[byte]Instruction)
	})
	var tests = []struct {
		code []byte
		gas  uint64
		expected
	}{
		{
			// Muliply 6 * 7 and return the result
			//
			// 60 06
			// 60 07
			// 02
			// 60 00
			// 53
			// 60 01
			// 60 00
			// f3
			code: HexToBytes("600660070260005360016000f3"),
			gas:  24,
			expected: expected{
				stack:      []*uint256.Int{},
				memory:     append([]byte{42}, zeroWord...)[:32],
				returndata: []byte{42},
				gasLeft:    1,
			},
		},
		{
			// Muliply 6 * 7 and return the result
			//
			// 60 02
			// 60 04
			// 04
			code: HexToBytes("6002600404"),
			gas:  11,
			expected: expected{
				stack:      []*uint256.Int{uint256.NewInt(2)},
				memory:     []byte{},
				returndata: []byte{},
				gasLeft:    0,
			},
		},
		{
			code: HexToBytes("600660070260005360016000f3"),
			gas:  13,
			expected: expected{
				// todo: ideally, in case of OutOfGas stack, memory, storage should all revert
				stack:      []*uint256.Int{uint256.NewInt(42)},
				memory:     []byte{},
				returndata: []byte{},
				gasLeft:    0,
			},
		},
		{
			// SSTORE and SLOAD
			//
			// 60 01
			// 60 00
			// 55
			// 60 00
			// 54
			code: HexToBytes("6001600055600054"),
			gas:  60, // should go out of gas
			expected: expected{
				stack:      []*uint256.Int{uint256.NewInt(1)},
				memory:     []byte{},
				returndata: []byte{},
				storage:    map[uint256.Int]*uint256.Int{*uint256.NewInt(0): uint256.NewInt(1)},
				gasLeft:    1,
			},
		},
		{
			// infinite loop that jumps back to the start everytime
			//
			// 5b
			// 60 00
			// 56
			code: HexToBytes("5b600056"),
			gas:  13, // should go out of gas
			expected: expected{
				stack:      []*uint256.Int{},
				memory:     []byte{},
				returndata: []byte{},
				gasLeft:    0,
			},
		},
		{
			// calculate 4^2
			//
			// 60 04
			// 80
			// 60 00
			// 5b
			// 81
			// 60 12
			// 57
			// 60 00
			// 53
			// 60 01
			// 60 00
			// f3
			// 5b
			// 82
			// 01
			// 90
			// 60 01
			// 90
			// 03
			// 90
			// 60 05
			// 56
			code: HexToBytes("60048060005b8160125760005360016000f35b8201906001900390600556"),
			gas:  250, // should go out of gas
			expected: expected{
				stack:      []*uint256.Int{uint256.NewInt(4), uint256.NewInt(0)},
				memory:     append([]byte{16}, zeroWord...)[0:32],
				returndata: []byte{16},
				gasLeft:    12,
			},
		},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%X", tt.code)
		t.Run(testname, func(t *testing.T) {
			ectx := NewExecutionCtx(
				tt.code,
				NewCalldata("abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"),
				NewStack(),
				NewMemory(),
				NewStorage(),
				tt.gas,
			)
			Run(ectx)
			assert.Equal(t, tt.expected.gasLeft, ectx.Gas)
			assert.Equal(t, tt.expected.stack, ectx.Stack.data)
			assert.Equal(t, tt.expected.memory, ectx.Memory.data)
			assert.Equal(t, tt.expected.returndata, ectx.Returndata)
		})
	}
}

func TestRunFailure(t *testing.T) {
	Init()
	t.Cleanup(func() {
		InstructionSet = make(map[byte]Instruction)
	})
	var tests = []struct {
		code []byte
		gas  uint64
	}{
		{
			// invalid opcode
			code: HexToBytes("8d"),
			gas:  10,
		},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%X", tt.code)
		t.Run(testname, func(t *testing.T) {
			ectx := NewExecutionCtx(
				tt.code,
				NewCalldata("abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"),
				NewStack(),
				NewMemory(),
				NewStorage(),
				tt.gas,
			)
			assert.PanicsWithError(
				t,
				fmt.Sprintf("inst not found for opcode %d", tt.code[0]),
				func() { Run(ectx) },
			)

		})
	}
}

func TestRunFailureInvalidPC(t *testing.T) {
	var tests = []struct {
		code []byte
		gas  uint64
	}{
		{
			// invalid opcode
			code: HexToBytes("00"),
			gas:  10,
		},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%X", tt.code)
		t.Run(testname, func(t *testing.T) {
			ectx := NewExecutionCtx(
				tt.code,
				NewCalldata("abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"),
				NewStack(),
				NewMemory(),
				NewStorage(),
				tt.gas,
			)
			// set and invalid value of pc such that is exceeds the length of the code
			ectx.pc = 5
			assert.PanicsWithError(
				t,
				"cannot find STOP OPCODE in Instruction Set",
				func() { Run(ectx) },
			)
		})
	}
}

func TestHexToBytes(t *testing.T) {
	var tests = []struct {
		hex   string
		bytes []byte
	}{
		{
			hex:   "01",
			bytes: []byte{1},
		},
		{
			hex:   "ff",
			bytes: []byte{255},
		},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%x", tt.hex)
		t.Run(testname, func(t *testing.T) {
			assert.Equal(t, tt.bytes, HexToBytes(tt.hex))
		})
	}

	assert.PanicsWithError(
		t,
		"encoding/hex: odd length hex string",
		func() {
			HexToBytes("3")
		},
	)

}
