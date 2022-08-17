package evm

import (
	"context"
	"errors"
	"fmt"
)

type ExecutionCtx struct {
	code       []byte
	pc         uint64
	Stack      *Stack
	Memory     *Memory
	Storage    *Storage
	Calldata   Calldata
	Returndata []byte
	Jumpdests  map[uint64]uint64
	Gas        uint64
	context    context.Context
	cancel     context.CancelFunc
}

func NewExecutionCtx(context context.Context,
	cancel context.CancelFunc,
	code []byte,
	stack *Stack,
	memory *Memory,
	storage *Storage,
	gas uint64) *ExecutionCtx {
	return &ExecutionCtx{
		context:    context,
		cancel:     cancel,
		code:       code,
		pc:         0,
		Stack:      stack,
		Memory:     memory,
		Storage:    storage,
		Returndata: make([]byte, 0),
		Jumpdests:  make(map[uint64]uint64),
		Gas:        gas,
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

// UseGas deducts the avialble gas. If the available gas is
// fully exhausted then it returns false
func (ectx *ExecutionCtx) UseGas(gas uint64) bool {
	// make sure that uint64 doesn't overflow
	if gas > ectx.Gas {
		ectx.Gas = 0
		return false
	}

	ectx.Gas -= gas
	return true
}

// Run starts the execution of the bytecode in the VM
func Run(ectx *ExecutionCtx) ([]byte, error) {

	ectx.ValidJumpDestination()
	fmt.Printf("set valid jump destination %v \n", ectx.Jumpdests)

	for {
		select {
		case <-ectx.context.Done():
			fmt.Println("execution complete", ectx.Gas)
			ectx.cancel()
			return ectx.Returndata, nil
		default:
			pcBefore := ectx.pc
			inst := decodeOpcode(ectx)
			// deduct gas from the budget before executing
			if ok := ectx.UseGas(inst.constantGas); !ok {
				// without gas we can't proceed
				ectx.cancel()
				return nil, errors.New("out of gas")
			}

			inst.executeFn(ectx)
			fmt.Printf("%s @ pc=%d\n", inst.name, pcBefore)
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
			ctx.Jumpdests[uint64(i)] = uint64(i)
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
// according to the given by offset and length of the data
// to be returned by the evm
func (ctx *ExecutionCtx) SetReturnData(offset, length uint64) {
	ctx.Returndata = ctx.Memory.LoadRange(offset, length)
	ctx.cancel()
}

// SetProgramCounter sets the PC in the execution context
func (ctx *ExecutionCtx) SetProgramCounter(pc uint64) {
	ctx.pc = pc
}
