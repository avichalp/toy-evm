package evm

import (
	"context"
	"fmt"
)

type ExecutionCtx struct {
	code       []byte
	pc         uint64
	stack      *Stack
	memory     *Memory
	calldata   Calldata
	returndata []byte
	jumpdests  map[uint64]uint64
	gas        uint64
	context    context.Context
	cancel     context.CancelFunc
}

func NewExecutionCtx(context context.Context, cancel context.CancelFunc, code []byte, stack *Stack, memory *Memory, gas uint64) *ExecutionCtx {
	return &ExecutionCtx{
		context:    context,
		cancel:     cancel,
		code:       code,
		pc:         0,
		stack:      stack,
		memory:     memory,
		returndata: make([]byte, 0),
		jumpdests:  make(map[uint64]uint64),
		gas:        gas,
	}
}

// decodeOpcode decodes the bytecode @ PC using
// the InstructionSet
func decodeOpcode(ctx *ExecutionCtx) Instruction {
	fmt.Println("decoding opcode")
	// Yellow paper section 9.4.1 (Machine State)
	if ctx.pc >= uint64(len(ctx.code)) {
		inst, ok := InstructionSet[0]
		if !ok {
			panic("Cannot find STOP OPCODE in Instruction Set")
		}
		return inst
	}

	opcode := ctx.ReadCode(1)
	fmt.Println("finding instruction for opcode", opcode)
	inst, ok := InstructionSet[opcode]
	if !ok {
		panic(fmt.Errorf("inst not found for opcode %d", opcode))

	}
	return inst
}

func (ectx *ExecutionCtx) useGas(gas uint64) {
	// make sure that uint64 doesn't overflow
	if gas > ectx.gas {
		ectx.gas = 0
	} else {
		ectx.gas -= gas
	}
}

// Run starts the execution of the bytecode in the VM
func Run(ectx *ExecutionCtx) {

	ectx.ValidJumpDestination()
	fmt.Printf("set valid jump destination %v \n", ectx.jumpdests)

	for {
		select {
		case <-ectx.context.Done():
			fmt.Println("execution complete", ectx.gas)
			ectx.cancel()
			return
		default:
			// no. of execution check
			if ectx.gas <= 0 {
				fmt.Println("out of gas", ectx.gas)
				ectx.cancel()
				return
			}

			pcBefore := ectx.pc
			inst := decodeOpcode(ectx)
			// deduct gas from the budget before executing
			ectx.useGas(inst.constantGas)
			inst.executeFn(ectx)

			fmt.Printf("%s @ pc=%d\n", inst.name, pcBefore)
			fmt.Printf("stack: %v\n", ectx.stack.stack)
			fmt.Printf("memory: %v\n", ectx.memory.memory)
			fmt.Printf("returndata: %v\n", ectx.returndata)
			fmt.Printf("gas left: %v\n", ectx.gas)
			fmt.Printf("\n")
		}
	}
}

// Stop stops the execution of the bytecode in the VM
func (ctx *ExecutionCtx) Stop() {
	ctx.cancel()
}

// ValidJumpDestination iterates over the bytecode.
// If it finds a valid destination, that is when the
// current byte == 0x5B, it remembers this index in a
// set (implemented as a hashmap).
//
// If the byte represents PUSH1-PUSH32, the index
// must increment by the number of bytes PUSH-N
// instruction will read from the code buffer plus one.
//
// For PUSH1 we only need to increment the index by 2.
// One time for the PUSH1 instruction and second time for
// the following 1 Byte. Similarly for PUSH2, index will
// increment by: 1 + (0x61-0x60 + 1)
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
	ctx.pc += numBytes
	return codeSegment[0]

}

// SetReturnData sets the return data into the memory region
// according to the given by offset and lenght of the data
// to be returned
func (ctx *ExecutionCtx) SetReturnData(offset, length uint64) {
	ctx.returndata = ctx.memory.LoadRange(offset, length)
	ctx.cancel()
}

// SetProgramCounter sets the PC in the execution context
func (ctx *ExecutionCtx) SetProgramCounter(pc uint64) {
	ctx.pc = pc
}
