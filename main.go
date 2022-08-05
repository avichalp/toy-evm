package main

import (
	"context"
	"flag"
	"fmt"
	"math"

	"golang.org/x/exp/slices"

	"github.com/holiman/uint256"
)

type Stack struct {
	stack    []uint256.Int
	maxDepth int
}

func NewStack() *Stack {
	s := make([]uint256.Int, 0)
	return &Stack{
		stack:    s,
		maxDepth: 1024,
	}
}

func invalidWord(word uint256.Int) bool {
	min := uint256.NewInt(0)
	max := uint256.NewInt(uint64(math.Pow(2, 256) - 1))
	return word.Lt(min) || word.Gt(max)
}

func (s *Stack) Push(item uint256.Int) {
	if invalidWord(item) {
		panic("Stack item too big")
	}
	if len(s.stack)+1 > s.maxDepth {
		panic("Stack Overflow")
	}
	s.stack = append(s.stack, item)
}

func (s *Stack) Pop() (item uint256.Int) {
	if len(s.stack) == 0 {
		panic("Stack underflow")
	}
	item = s.stack[len(s.stack)-1]
	slices.Delete(s.stack, len(s.stack)-1, len(s.stack))
	return
}

type Memory struct {
	memory []uint8
}

func NewMemory() *Memory {
	m := make([]uint8, 0)
	return &Memory{memory: m}
}

func (m *Memory) Store(offset uint64, value uint8) {
	/* if offset < 0 || offset > uint64(math.Pow(2, 256))-1 {
		panic(fmt.Sprintf("Invalid memory access %d", offset))
	} */
	if value < 0 || value > math.MaxUint8 {
		panic(fmt.Sprintf("Invalid memory value %d", value))
	}

	// expand if needed
	if offset >= uint64(len(m.memory)) {
		for i := 0; i < int(offset-uint64(len(m.memory))+1); i++ {
			m.memory = append(m.memory, 0)
		}
	}
	m.memory[offset] = value
}

func (m *Memory) Load(offset uint64) uint8 {
	if offset >= uint64(len(m.memory)) {
		return 0
	}
	return m.memory[offset]
}

func decodeOpcode(ctx context.Context) context.Context {
	fmt.Println("decoding opcode")
	return ctx
}

func execute(numSteps *int, cancel context.CancelFunc) {
	// todo: if number of instruction exceeded or excecution is complete
	if *numSteps >= 3 {
		fmt.Println("NUM OF STEPS REACHED")
		cancel()
		return
	}
	fmt.Printf("exectuing opcode\n")
	*numSteps++
}

func run(code, calldata string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	numSteps := 0

	for {
		select {
		case <-ctx.Done():
			fmt.Println("NUM STEPS EXECUTED", numSteps)
			cancel()
			return
		default:
			decodeOpcode(ctx)
			execute(&numSteps, cancel)
		}
	}
}

func main() {
	var code string
	var calldata string
	flag.StringVar(&code, "code", "0x0", "hex data of the code to run")
	flag.StringVar(&calldata, "calldata", "0x0", "hex data to use as input")
	flag.Parse()
	fmt.Printf("code: %s, calldata %s\n", code, calldata)
	run(code, calldata)

}
