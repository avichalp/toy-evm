package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math"

	"golang.org/x/exp/slices"

	"github.com/holiman/uint256"
	"github.com/status-im/keycard-go/hexutils"
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

func decodeOpcode(ctx *ExecutionCtx) (*Instruction, error) {
	fmt.Println("decoding opcode")
	if ctx.pc < 0 || ctx.pc > len(ctx.code) {
		panic(fmt.Sprintf("invalid code offset: %s %d\n", ctx.code, ctx.pc))
	}

	opcode := ctx.ReadCode(1)
	if inst, ok := InstructionByOpcode[opcode]; ok {
		return inst, nil
	}

	return nil, errors.New("")
}

func _execute(numSteps *int, cancel context.CancelFunc) {
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
	// ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	//numSteps := 0
	stack, memory := NewStack(), NewMemory()

	ectx := NewExecutionCtx(hexutils.HexToBytes(code), stack, memory)

	for !ectx.stopped {
		pcBefore := ectx.pc
		if inst, err := decodeOpcode(ectx); err == nil {
			inst.executeFn(ectx)
			fmt.Printf("%v @ pc=%d\n", inst, pcBefore)
			fmt.Printf("%v\n", ectx)
			fmt.Printf("\n")
		}

		/* select {
		case <-ctx.Done():
			fmt.Println("NUM STEPS EXECUTED", numSteps)
			cancel()
			return
		default:
			decodeOpcode(ectx)
			_execute(&numSteps, cancel)
		} */
	}
}

type ExecutionCtx struct {
	code    []byte
	pc      int
	stack   *Stack
	memory  *Memory
	stopped bool
}

func NewExecutionCtx(code []byte, stack *Stack, memory *Memory) *ExecutionCtx {
	return &ExecutionCtx{
		code:    code,
		pc:      0,
		stack:   stack,
		memory:  memory,
		stopped: false,
	}
}

func (ctx *ExecutionCtx) Stop() {
	ctx.stopped = true
}

// ReadCode returns the next numBytes from the code
// buffer as an integer and advances pc by numBytes
func (ctx *ExecutionCtx) ReadCode(numBytes int) *uint256.Int {
	codeSegment := ctx.code[ctx.pc : ctx.pc+numBytes]
	//codeSegmentHex := hexutils.BytesToHex(codeSegment)
	codeHex := fmt.Sprintf("0x%x", ctx.code)
	codeSegmentHex := fmt.Sprintf("0x%x", codeSegment)
	fmt.Printf("reading code: %s, segment: %s\n", codeHex, codeSegmentHex)
	value, err := uint256.FromHex(codeSegmentHex)
	if err != nil {
		panic(err)
	}
	// increment the program counter
	ctx.pc += numBytes
	return value
}

type ExecuteFn func(*ExecutionCtx)
type Instruction struct {
	opcode    *uint256.Int
	name      string
	executeFn ExecuteFn
}

var Instructions = make([]*Instruction, 0)
var InstructionByOpcode = make(map[*uint256.Int]*Instruction)

func RegisterInstruction(opcode *uint256.Int, name string, executeFn ExecuteFn) *Instruction {
	inst := &Instruction{
		opcode:    opcode,
		name:      name,
		executeFn: executeFn,
	}
	Instructions = append(Instructions, inst)

	InstructionByOpcode[opcode] = inst
	fmt.Printf("registering %s: %v\n", opcode, inst)
	return inst
}

func main() {
	var code string
	var calldata string
	flag.StringVar(&code, "code", "0x0", "hex data of the code to run")
	flag.StringVar(&calldata, "calldata", "0x0", "hex data to use as input")
	flag.Parse()
	fmt.Printf("code: %s, calldata %s\n", code, calldata)

	if stopOP, err := uint256.FromHex("0x00"); err == nil {
		RegisterInstruction(stopOP, "STOP", func(ctx *ExecutionCtx) { ctx.Stop() })
	}

	if pushOP, err := uint256.FromHex("0x60"); err == nil {
		RegisterInstruction(
			pushOP,
			"PUSH1",
			func(ctx *ExecutionCtx) {
				ctx.stack.Push(ctx.ReadCode(1))
			})
	}

	if addOP, err := uint256.FromHex("0x01"); err == nil {
		RegisterInstruction(
			addOP,
			"ADD",
			func(ctx *ExecutionCtx) {
				op1, op2 := ctx.stack.Pop(), ctx.stack.Pop()
				result := uint256.NewInt(0)
				mod := uint256.NewInt(uint64(math.Pow(2, 256)))
				result.AddMod(op1, op2, mod)
				ctx.stack.Push(result)
			})
	}
	run(code, calldata)
}
