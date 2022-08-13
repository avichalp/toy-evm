package evm

import (
	"fmt"
	"math"

	"github.com/holiman/uint256"
)

type ExecuteFn func(*ExecutionCtx)

type Instruction struct {
	opcode    byte
	name      string
	executeFn ExecuteFn
}

var InstructionByOpcode map[byte]Instruction

func Init() {
	InstructionByOpcode = map[byte]Instruction{
		0x0:  {0x0, "STOP", opStop},
		0x01: {0x01, "ADD", opAdd},
		0x02: {0x02, "MUL", opMul},
		0x03: {0x03, "SUB", opSub},
		0x60: {0x60, "PUSH1", opPush1},
		0xF3: {0xF3, "RETURN", opReturn},
		0x56: {0x56, "JUMP", opJump},
		0x57: {0x57, "JUMPI", opJumpi},
		0x51: {0x51, "MLOAD", opMload},
		0x52: {0x52, "MSTORE", opMstore},
		0x53: {0x53, "MSTORE8", opMstore8},
		0x58: {0x58, "PC", opProgramCounter},
		0x59: {0x59, "MSIZE", opMsize},
		0x5B: {0x5B, "JUMPDEST", opJumpdest},
		0x80: {0x80, "DUP1", opDup1},
		0x81: {0x81, "DUP2", opDup2},
		0x82: {0x82, "DUP3", opDup3},
		0x90: {0x90, "SWAP1", OpSwap1},
		0x35: {0x35, "CALLDATALOAD", opCallDataLoad},
	}

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
