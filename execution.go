package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/holiman/uint256"
	"github.com/status-im/keycard-go/hexutils"
)

func decodeOpcode(ctx *ExecutionCtx) (*Instruction, error) {
	fmt.Println("decoding opcode")
	if ctx.pc < 0 {
		panic(fmt.Sprintf("invalid code offset: code: %v, pc: %d\n", ctx.code, ctx.pc))
	}
	// Yellow paper section 9.4.1 (Machine State)
	if ctx.pc >= len(ctx.code) {
		inst, ok := InstructionByOpcode[*uint256.NewInt(0)]
		if !ok {
			panic("Cannot find STOP OPCODE in Instruction Registry")
		}
		return inst, nil
	}

	opcode := ctx.ReadCode(1)
	fmt.Println("finding instruction for opcode", opcode)
	inst, ok := InstructionByOpcode[*opcode]
	if !ok {
		return nil, fmt.Errorf("inst not found for opcode %d", opcode)

	}
	fmt.Println("Instruction matched", inst)
	return inst, nil
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
		inst, err := decodeOpcode(ectx)
		if err != nil {
			panic(err)
		}

		inst.executeFn(ectx)
		fmt.Printf("%s @ pc=%d\n", inst.name, pcBefore)
		fmt.Printf("stack: %v\n", ectx.stack.stack)
		fmt.Printf("memory: %v\n", ectx.memory.memory)
		fmt.Printf("return: %v\n", ectx.returndata)
		fmt.Printf("\n")

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
	code       []byte
	pc         int
	stack      *Stack
	memory     *Memory
	stopped    bool
	returndata []byte
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
	codeSegmentHex := hexutils.BytesToHex(codeSegment)

	codeHex := fmt.Sprintf("0x%x", ctx.code)
	fmt.Printf("reading code: %s, bytes: %d, segment: %s\n", codeHex, numBytes, codeSegmentHex)

	// removing leading zeros
	codeSegmentHex = strings.TrimLeft(codeSegmentHex, "0")

	// increment the program counter
	ctx.pc += numBytes

	// default hex -> decimal conversion for 0
	if codeSegmentHex == "" {
		return uint256.NewInt(0)
	}

	value, err := uint256.FromHex("0x" + codeSegmentHex)
	if err != nil {
		panic(err)
	}
	return value

}

func (ctx *ExecutionCtx) SetReturnData(offset, length uint64) {
	ctx.stopped = true
	ctx.returndata = ctx.memory.LoadRange(offset, length)
}

type ExecuteFn func(*ExecutionCtx)
type Instruction struct {
	opcode    *uint256.Int
	name      string
	executeFn ExecuteFn
}
