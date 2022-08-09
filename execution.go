package main

import (
	"context"
	"fmt"

	"github.com/status-im/keycard-go/hexutils"
)

func decodeOpcode(ctx *ExecutionCtx) (*Instruction, error) {
	fmt.Println("decoding opcode")
	// Yellow paper section 9.4.1 (Machine State)
	if ctx.pc >= uint64(len(ctx.code)) {
		inst, ok := InstructionByOpcode[0]
		if !ok {
			panic("Cannot find STOP OPCODE in Instruction Registry")
		}
		return inst, nil
	}

	opcode := ctx.ReadCode(1)
	fmt.Println("finding instruction for opcode", opcode)
	inst, ok := InstructionByOpcode[opcode]
	if !ok {
		return nil, fmt.Errorf("inst not found for opcode %d", opcode)

	}
	fmt.Println("Instruction matched", inst)
	return inst, nil
}

func run(code, calldata string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stack, memory := NewStack(), NewMemory()
	ectx := NewExecutionCtx(ctx, cancel, hexutils.HexToBytes(code), stack, memory)

	ectx.ValidJumpDestination()
	fmt.Printf("set valid jump destination %v \n", ectx.jumpdests)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("execution complete", ectx.numSteps)
			cancel()
			return
		default:
			// no. of execution check
			if ectx.numSteps >= 3 {
				fmt.Println("out of gas", ectx.numSteps)
				cancel()
				return
			}

			pcBefore := ectx.pc
			inst, err := decodeOpcode(ectx)
			if err != nil {
				panic(err)
			}

			inst.executeFn(ectx)
			ectx.numSteps++
			fmt.Printf("%s @ pc=%d\n", inst.name, pcBefore)
			fmt.Printf("stack: %v\n", ectx.stack.stack)
			fmt.Printf("memory: %v\n", ectx.memory.memory)
			fmt.Printf("return: %v\n", ectx.returndata)
			fmt.Printf("\n")
		}
	}
}

type ExecutionCtx struct {
	code       []byte
	pc         uint64
	stack      *Stack
	memory     *Memory
	stopped    bool
	returndata []byte
	jumpdests  map[uint64]uint64
	numSteps   int
	context    context.Context
	cancel     context.CancelFunc
}

func NewExecutionCtx(context context.Context, cancel context.CancelFunc, code []byte, stack *Stack, memory *Memory) *ExecutionCtx {
	return &ExecutionCtx{
		context:   context,
		cancel:    cancel,
		code:      code,
		pc:        0,
		stack:     stack,
		memory:    memory,
		stopped:   false,
		jumpdests: make(map[uint64]uint64),
		numSteps:  0,
	}
}

func (ctx *ExecutionCtx) Stop() {
	ctx.cancel()
}

func (ctx *ExecutionCtx) ValidJumpDestination() {
	i := 0
	for i < len(ctx.code) {
		currentOP := ctx.code[i]
		if currentOP == 0x5B { // OPCODE of JUMPDEST
			ctx.jumpdests[uint64(i)] = uint64(i)
		} else if currentOP >= 0x60 && currentOP <= 0x7F {
			i += int(currentOP) - 0x60 + 1
		}
		i += 1
	}
}

// ReadCode returns the next numBytes from the code
// buffer as an integer and advances pc by numBytes
func (ctx *ExecutionCtx) ReadCode(numBytes uint64) byte {
	codeSegment := ctx.code[ctx.pc : ctx.pc+numBytes]
	codeHex := fmt.Sprintf("0x%x", ctx.code)
	fmt.Printf("reading code: %s, bytes: %d, segment: %s\n", codeHex, numBytes, codeSegment)

	// increment the program counter
	ctx.pc += numBytes

	return codeSegment[0]

}

func (ctx *ExecutionCtx) SetReturnData(offset, length uint64) {
	ctx.returndata = ctx.memory.LoadRange(offset, length)
	ctx.cancel()
}

func (ctx *ExecutionCtx) SetProgramCounter(pc uint64) {
	ctx.pc = pc
}

type ExecuteFn func(*ExecutionCtx)
type Instruction struct {
	opcode    byte
	name      string
	executeFn ExecuteFn
}
