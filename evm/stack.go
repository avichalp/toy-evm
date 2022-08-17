package evm

import (
	"fmt"
	"math"

	"github.com/holiman/uint256"
	"golang.org/x/exp/slices"
)

type Stack struct {
	data     []*uint256.Int
	maxDepth int
}

func NewStack() *Stack {
	s := make([]*uint256.Int, 0)
	return &Stack{
		data:     s,
		maxDepth: 1024,
	}
}

func invalidWord(word *uint256.Int) bool {
	min := uint256.NewInt(0)
	max := uint256.NewInt(uint64(math.Pow(2, 256) - 1))
	return word.Lt(min) || word.Gt(max)
}

func (s *Stack) Push(item *uint256.Int) {
	if invalidWord(item) {
		panic("Stack item too big")
	}
	if len(s.data)+1 > s.maxDepth {
		panic("Stack Overflow")
	}
	s.data = append(s.data, item)
}

func (s *Stack) Pop() (item *uint256.Int) {
	if len(s.data) == 0 {
		panic("Stack underflow")
	}
	item = s.data[len(s.data)-1]
	s.data = slices.Delete(s.data, len(s.data)-1, len(s.data))
	return
}

// Peek returns a stack element without popping it
// eg: Peek(0) will return the top of the stack
func (s *Stack) Peek(i uint16) (item *uint256.Int) {
	// there can only be 1024 stack frames
	length := uint16(len(s.data))
	if length < i {
		panic(fmt.Errorf("stack underflow %d", i))
	}

	return s.data[length-1-i]
}

// Swap the top of the stack with the i+1th element
func (s *Stack) Swap(i uint16) {
	if i == 0 {
		return
	}
	length := uint16(len(s.data))

	if length < i {
		panic(fmt.Errorf("stack underflow %d", i))
	}

	s.data[length-1], s.data[length-1-i] = s.data[length-1-i], s.data[length-1]
}

func (s *Stack) String() string {
	return fmt.Sprintf("stack: %s", s.data)
}
