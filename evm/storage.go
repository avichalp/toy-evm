package evm

import (
	"fmt"

	"github.com/holiman/uint256"
)

type Storage struct {
	data map[uint256.Int]*uint256.Int
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[uint256.Int]*uint256.Int),
	}
}

func (s *Storage) get(slot uint256.Int) *uint256.Int {
	value, ok := s.data[slot]
	if !ok {
		fmt.Println("cannot find slot in storage", slot)
		return uint256.NewInt(0)
	}

	return value
}

func (s *Storage) put(slot *uint256.Int, value *uint256.Int) {
	s.data[*slot] = value
}
