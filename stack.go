package main

import (
	"math"

	"github.com/holiman/uint256"
	"golang.org/x/exp/slices"
)

type Stack struct {
	stack    []*uint256.Int
	maxDepth int
}

func NewStack() *Stack {
	s := make([]*uint256.Int, 0)
	return &Stack{
		stack:    s,
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
	if len(s.stack)+1 > s.maxDepth {
		panic("Stack Overflow")
	}
	s.stack = append(s.stack, item)
}

func (s *Stack) Pop() (item *uint256.Int) {
	if len(s.stack) == 0 {
		panic("Stack underflow")
	}
	item = s.stack[len(s.stack)-1]
	s.stack = slices.Delete(s.stack, len(s.stack)-1, len(s.stack))
	return
}
