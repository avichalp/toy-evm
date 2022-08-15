package evm

import (
	"context"
	"fmt"
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

type expected struct {
	stack      []*uint256.Int
	memory     []byte
	returndata []byte
	gasLeft    uint64
}

func TestRunSucess(t *testing.T) {
	Init()
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
				memory:     []byte{42},
				returndata: []byte{42},
				gasLeft:    1,
			},
		},
		{
			code: HexToBytes("600660070260005360016000f3"),
			gas:  13,
			expected: expected{
				stack:      []*uint256.Int{uint256.NewInt(42), uint256.NewInt(0)},
				memory:     []byte{},
				returndata: []byte{},
				gasLeft:    0,
			},
		},
		{
			// infinite loop that jumps back to the start everytime
			//
			// 5b
			// 60 00
			// 56
			code: HexToBytes("5b600056"),
			gas:  5, // should go out of gas
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
				memory:     []byte{16},
				returndata: []byte{16},
				gasLeft:    12,
			},
		},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%X", tt.code)
		t.Run(testname, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			ectx := NewExecutionCtx(
				ctx,
				cancel,
				tt.code,
				NewStack(),
				NewMemory(),
				tt.gas,
			)
			Run(ectx)
			assert.Equal(t, tt.expected.gasLeft, ectx.gas)
			assert.Equal(t, tt.expected.stack, ectx.stack.stack)
			assert.Equal(t, tt.expected.memory, ectx.memory.memory)
			assert.Equal(t, tt.expected.returndata, ectx.returndata)
		})
	}
}
