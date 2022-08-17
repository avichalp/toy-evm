package evm

import (
	"fmt"
	"strings"

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

func (s *Storage) Get(slot uint256.Int) *uint256.Int {
	value, ok := s.data[slot]
	if !ok {
		fmt.Println("cannot find slot in storage", slot)
		return uint256.NewInt(0)
	}

	return value
}

func (s *Storage) Put(slot *uint256.Int, value *uint256.Int) {
	s.data[*slot] = value
}

func (s *Storage) String() string {
	strs := []string{"storage: \n"}
	for k, v := range s.data {
		strs = append(strs, fmt.Sprintf("%d: %d\n", k, v))
	}
	return strings.Join(strs, "")
}
