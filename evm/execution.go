package evm

import (
	"context"
	"fmt"
	"math"

	"github.com/holiman/uint256"
)

var Instructions = make([]*Instruction, 0)
var InstructionByOpcode = make(map[byte]*Instruction)

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
			inst, err := decodeOpcode(ectx)
			if err != nil {
				panic(err)
			}

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

type ExecuteFn func(*ExecutionCtx)

type Instruction struct {
	opcode    byte
	name      string
	executeFn ExecuteFn
}

func RegisterInstruction(opcode byte, name string, executeFn ExecuteFn) *Instruction {
	inst := &Instruction{
		opcode:    opcode,
		name:      name,
		executeFn: executeFn,
	}
	Instructions = append(Instructions, inst)

	InstructionByOpcode[opcode] = inst
	fmt.Printf("registering %d: %v\n", opcode, inst)
	return inst
}

func opStop(ctx *ExecutionCtx) { ctx.Stop() }

func opPush1(ctx *ExecutionCtx) {
	ctx.stack.Push(uint256.NewInt(uint64(ctx.ReadCode(1))))
}

func opAdd(ctx *ExecutionCtx) {
	op1, op2 := ctx.stack.Pop(), ctx.stack.Pop()
	result := uint256.NewInt(0)
	mod := uint256.NewInt(uint64(math.Pow(2, 256)))
	result.AddMod(op1, op2, mod)
	ctx.stack.Push(result)
}

func opMul(ctx *ExecutionCtx) {
	op1, op2 := ctx.stack.Pop(), ctx.stack.Pop()
	result := uint256.NewInt(0)
	mod := uint256.NewInt(uint64(math.Pow(2, 256)))
	result.MulMod(op1, op2, mod)
	ctx.stack.Push(result)
}

func opSub(ctx *ExecutionCtx) {
	op1, op2 := ctx.stack.Pop(), ctx.stack.Pop()
	result := uint256.NewInt(0)
	mod := uint256.NewInt(uint64(math.Pow(2, 256)))
	result.Sub(op1, op2)
	result.Mod(result, mod)
	ctx.stack.Push(result)
}

func opReturn(ctx *ExecutionCtx) {
	op1, op2 := ctx.stack.Pop(), ctx.stack.Pop()
	// TODO: these arguments should be uint256
	ctx.SetReturnData(op1.Uint64(), op2.Uint64())
}

func opJump(ctx *ExecutionCtx) {
	pc := ctx.stack.Pop().Uint64()
	fmt.Printf("valid jump dests: %v\n", ctx.jumpdests)
	if _, ok := ctx.jumpdests[pc]; !ok {
		panic(fmt.Errorf("invalid jump destination %d", pc))
	}
	ctx.SetProgramCounter(pc)
}

func opJumpi(ctx *ExecutionCtx) {
	pc, cond := ctx.stack.Pop().Uint64(), ctx.stack.Pop()
	if cond.Cmp(uint256.NewInt(0)) != 0 {
		if _, ok := ctx.jumpdests[pc]; !ok {
			panic(fmt.Errorf("invalid jump destination %d", pc))
		}
		ctx.SetProgramCounter(pc)
	}
}

func opMload(ctx *ExecutionCtx) {
	offset := ctx.stack.Pop()
	ctx.stack.Push(ctx.memory.LoadWord(offset.Uint64()))
}

func opMstore8(ctx *ExecutionCtx) {
	op1, op2 := ctx.stack.Pop(), ctx.stack.Pop()
	// MSTORE8 pops an offset and a word from the stack,
	// and stores the lowest byte of that word in memory
	op2.Mod(op2, uint256.NewInt(256))
	ctx.memory.StoreByte(op1.Uint64(), uint8(op2.Uint64()))
}

func opMstore(ctx *ExecutionCtx) {
	offset, value := ctx.stack.Pop(), ctx.stack.Pop()
	ctx.memory.StoreWord(offset.Uint64(), *value)
}

func opProgramCounter(ctx *ExecutionCtx) {
	ctx.stack.Push(uint256.NewInt(ctx.pc))
}

func opMsize(ctx *ExecutionCtx) {
	ctx.stack.Push(uint256.NewInt(ctx.memory.ActiveWords() * 32))
}

func opJumpdest(_ *ExecutionCtx) {}

func opDup1(ctx *ExecutionCtx) {
	ctx.stack.Push(ctx.stack.Peek(0))
}

func opDup2(ctx *ExecutionCtx) {
	ctx.stack.Push(ctx.stack.Peek(1))
}

func opDup3(ctx *ExecutionCtx) {
	ctx.stack.Push(ctx.stack.Peek(2))
}

func OpSwap1(ctx *ExecutionCtx) {
	ctx.stack.Swap(1)
}

func opCallDataLoad(ctx *ExecutionCtx) {
	// geth limits the size of calldata to uint64
	// https://github.com/ethereum/go-ethereum/blob/440c9fcf75d9d5383b72646a65d5e21fa7ab6a26/core/vm/instructions.go
	if offset, overflow := ctx.stack.Pop().Uint64WithOverflow(); !overflow {
		ctx.stack.Push(ctx.calldata.ReadWord(offset))
	}
}

func Init() {
	RegisterInstruction(0x0, "STOP", opStop)
	RegisterInstruction(0x60, "PUSH1", opPush1)
	RegisterInstruction(0x01, "ADD", opAdd)
	RegisterInstruction(0x02, "MUL", opMul)
	RegisterInstruction(0x03, "SUB", opSub)
	RegisterInstruction(0xF3, "RETURN", opReturn)
	RegisterInstruction(0x56, "JUMP", opJump)
	RegisterInstruction(0x57, "JUMPI", opJumpi)
	RegisterInstruction(0x51, "MLOAD", opMload)
	RegisterInstruction(0x53, "MSTORE8", opMstore8)
	RegisterInstruction(0x52, "MSTORE", opMstore)
	RegisterInstruction(0x58, "PC", opProgramCounter)
	RegisterInstruction(0x59, "MSIZE", opMsize)
	RegisterInstruction(0x5B, "JUMPDEST", opJumpdest)
	RegisterInstruction(0x80, "DUP1", opDup1)
	RegisterInstruction(0x81, "DUP2", opDup2)
	RegisterInstruction(0x82, "DUP3", opDup3)
	RegisterInstruction(0x90, "SWAP1", OpSwap1)
	RegisterInstruction(0x35, "CALLDATALOAD", opCallDataLoad)
}
