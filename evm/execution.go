package evm

import (
	"context"
	"fmt"
)

func decodeOpcode(ctx *ExecutionCtx) Instruction {
	fmt.Println("decoding opcode")
	// Yellow paper section 9.4.1 (Machine State)
	if ctx.pc >= uint64(len(ctx.code)) {
		inst, ok := InstructionByOpcode[0]
		if !ok {
			panic("Cannot find STOP OPCODE in Instruction Registry")
		}
		return inst
	}

	opcode := ctx.ReadCode(1)
	fmt.Println("finding instruction for opcode", opcode)
	inst, ok := InstructionByOpcode[opcode]
	if !ok {
		panic(fmt.Errorf("inst not found for opcode %d", opcode))

	}
	return inst
}

func Run(ectx *ExecutionCtx) {

	ectx.ValidJumpDestination()
	fmt.Printf("set valid jump destination %v \n", ectx.jumpdests)

	for {
		select {
		case <-ectx.context.Done():
			fmt.Println("execution complete", ectx.steps)
			ectx.cancel()
			return
		default:
			// no. of execution check
			if ectx.steps < 0 {
				fmt.Println("out of gas", ectx.steps)
				ectx.cancel()
				return
			}

			pcBefore := ectx.pc
			inst := decodeOpcode(ectx)

			inst.executeFn(ectx)
			ectx.steps--
			fmt.Printf("%s @ pc=%d\n", inst.name, pcBefore)
			fmt.Printf("stack: %v\n", ectx.stack.stack)
			fmt.Printf("memory: %v\n", ectx.memory.memory)
			fmt.Printf("returndata: %v\n", ectx.returndata)
			fmt.Printf("\n")
		}
	}
}

type ExecutionCtx struct {
	code       []byte
	pc         uint64
	stack      *Stack
	memory     *Memory
	calldata   Calldata
	returndata []byte
	jumpdests  map[uint64]uint64
	steps      int
	context    context.Context
	cancel     context.CancelFunc
}

func NewExecutionCtx(context context.Context, cancel context.CancelFunc, code []byte, stack *Stack, memory *Memory, steps int) *ExecutionCtx {
	return &ExecutionCtx{
		context:    context,
		cancel:     cancel,
		code:       code,
		pc:         0,
		stack:      stack,
		memory:     memory,
		returndata: make([]byte, 0),
		jumpdests:  make(map[uint64]uint64),
		steps:      steps,
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
