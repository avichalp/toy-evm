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
	pc         uint64
	stack      *Stack
	memory     *Memory
	stopped    bool
	returndata []byte
	jumpdests  map[uint64]uint64
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

/* func (ctx *ExecutionCtx) ValidJumpDestination() map[uint64]uint64 {
	dests := make(map[uint64]uint64)
	i := 0
	for i < len(ctx.code) {
		currentOP := ctx.code[i]

	}
} */

// ReadCode returns the next numBytes from the code
// buffer as an integer and advances pc by numBytes
func (ctx *ExecutionCtx) ReadCode(numBytes uint64) byte {
	codeSegment := ctx.code[ctx.pc : ctx.pc+numBytes]
	// codeSegmentHex := hexutils.BytesToHex(codeSegment)

	codeHex := fmt.Sprintf("0x%x", ctx.code)
	fmt.Printf("reading code: %s, bytes: %d, segment: %s\n", codeHex, numBytes, codeSegment)

	// removing leading zeros
	// codeSegmentHex = strings.TrimLeft(codeSegmentHex, "0")

	// increment the program counter
	ctx.pc += numBytes

	// default hex -> decimal conversion for 0
	/* if codeSegmentHex == "" {
		return 0
	}
	*/
	/* value, err := uint256.FromHex("0x" + codeSegmentHex)
	if err != nil {
		panic(err)
	} */
	return codeSegment[0]

}

func (ctx *ExecutionCtx) SetReturnData(offset, length uint64) {
	ctx.stopped = true
	ctx.returndata = ctx.memory.LoadRange(offset, length)
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
