package evm

import (
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func TestPush(t *testing.T) {
	stack := NewStack()
	item := uint256.NewInt(1)
	stack.Push(item)
	assert.Equal(t, item, stack.data[0])

	// test overflow
	assert.PanicsWithError(t, "stack overflow", func() {
		for i := 0; i < 1024; i++ {
			stack.Push(uint256.NewInt(0))
		}
	})
}

func TestSwap(t *testing.T) {
	stack := NewStack()
	item1 := uint256.NewInt(1)
	stack.Push(item1)
	item2 := uint256.NewInt(2)
	stack.Push(item2)

	// test valid swap
	stack.Swap(1)
	assert.Equal(t, item1, stack.Peek(0))
	assert.Equal(t, item2, stack.Peek(1))

	// test no swap
	stack.Swap(0)
	assert.Equal(t, item1, stack.Peek(0))
	assert.Equal(t, item2, stack.Peek(1))

	// test invalid swap
	assert.PanicsWithError(t, "stack underflow 2", func() {
		stack.Swap(2)
	})

}

func TestPop(t *testing.T) {
	stack := NewStack()
	item := uint256.NewInt(1)
	stack.Push(item)

	// test normal pop
	assert.Equal(t, item, stack.Pop())

	// test underflow
	assert.PanicsWithError(t, "stack underflow", func() {
		stack.Pop()
	})
}

func TestPeek(t *testing.T) {
	stack := NewStack()
	item := uint256.NewInt(1)
	stack.Push(item)
	// test valid peek
	assert.Equal(t, item, stack.Peek(0))

	// test invalid peeks
	assert.PanicsWithError(t, "invalid peek index 1", func() {
		stack.Peek(1)
	})
	assert.Equal(t, item, stack.Pop())
	assert.PanicsWithError(t, "invalid peek index 1", func() {
		stack.Peek(1)
	})
	assert.PanicsWithError(t, "invalid peek index 0", func() {
		stack.Peek(0)
	})
}

func TestStackString(t *testing.T) {
	stack := NewStack()
	item1 := uint256.NewInt(1)
	stack.Push(item1)
	item2 := uint256.NewInt(2)
	stack.Push(item2)
	assert.Equal(t, "stack: [1 2]", stack.String())
}
