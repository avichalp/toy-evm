package evm

import (
	"fmt"

	"github.com/holiman/uint256"
)

type ExecuteFn func(*ExecutionCtx)

type Instruction struct {
	opcode      byte
	name        string
	executeFn   ExecuteFn
	constantGas uint64
	// todo: dynamicGas
}

var InstructionSet map[byte]Instruction

// see geth: core/vm/gas.go
// Gas costs
const (
	GasQuickStep   uint64 = 2
	GasFastestStep uint64 = 3
	GasFastStep    uint64 = 5
	GasMidStep     uint64 = 8
	GasSlowStep    uint64 = 10
	GasExtStep     uint64 = 20
)

func Init() {
	InstructionSet = map[byte]Instruction{
		0x0:  {0x0, "STOP", opStop, 0},
		0x01: {0x01, "ADD", opAdd, GasFastestStep},
		0x02: {0x02, "MUL", opMul, GasFastStep},
		0x03: {0x03, "SUB", opSub, GasFastestStep},
		0x04: {0x04, "DIV", opDiv, GasFastStep},
		0x08: {0x08, "ADDMOD", opAddMod, GasMidStep},
		0x09: {0x09, "MULMOD", opMulMod, GasMidStep},
		0x06: {0x06, "MOD", opMod, GasFastStep},
		0x60: {0x60, "PUSH1", opPush1, GasFastestStep},
		0xF3: {0xF3, "RETURN", opReturn, 0},
		0x56: {0x56, "JUMP", opJump, GasMidStep},
		0x57: {0x57, "JUMPI", opJumpi, GasSlowStep},
		0x51: {0x51, "MLOAD", opMload, GasFastestStep},
		0x52: {0x52, "MSTORE", opMstore, GasFastestStep},
		0x53: {0x53, "MSTORE8", opMstore8, GasFastestStep},
		0x54: {0x54, "SLOAD", opSload, 50},
		0x55: {0x55, "SSTORE", opSstore, 0},
		0x58: {0x58, "PC", opProgramCounter, GasQuickStep},
		0x59: {0x59, "MSIZE", opMsize, GasQuickStep},
		0x5a: {0x5a, "GAS", opGas, GasQuickStep},
		0x5B: {0x5B, "JUMPDEST", opJumpdest, 1},
		0x80: {0x80, "DUP1", opDup1, GasFastestStep},
		0x81: {0x81, "DUP2", opDup2, GasFastestStep},
		0x82: {0x82, "DUP3", opDup3, GasFastestStep},
		0x90: {0x90, "SWAP1", OpSwap1, GasFastestStep},
		0x35: {0x35, "CALLDATALOAD", opCalldataLoad, GasFastestStep},
		0x36: {0x36, "CALLDATASIZE", opCalldataSize, GasQuickStep},
		0x38: {0x38, "CODESIZE", opCodeSize, GasQuickStep},
	}

}

func opStop(ctx *ExecutionCtx) { ctx.Stop() }

func opPush1(ctx *ExecutionCtx) {
	ctx.Stack.Push(uint256.NewInt(uint64(ctx.ReadCode(1))))
}

func opAdd(ctx *ExecutionCtx) {
	op1, op2 := ctx.Stack.Pop(), ctx.Stack.Pop()
	result := uint256.NewInt(0)
	result.Add(op1, op2)
	ctx.Stack.Push(result)
}

func opMul(ctx *ExecutionCtx) {
	op1, op2 := ctx.Stack.Pop(), ctx.Stack.Pop()
	result := uint256.NewInt(0)
	result.Mul(op1, op2)
	ctx.Stack.Push(result)
}

func opSub(ctx *ExecutionCtx) {
	op1, op2 := ctx.Stack.Pop(), ctx.Stack.Pop()
	result := uint256.NewInt(0)
	result.Sub(op1, op2)
	ctx.Stack.Push(result)
}

func opDiv(ctx *ExecutionCtx) {
	n, d := ctx.Stack.Pop(), ctx.Stack.Pop()
	result := uint256.NewInt(0)
	result.Div(n, d)
	ctx.Stack.Push(result)
}

func opMod(ctx *ExecutionCtx) {
	op1, op2 := ctx.Stack.Pop(), ctx.Stack.Pop()
	// if op2 equals 0 the output is 0
	result := uint256.NewInt(0)
	result.Mod(op1, op2)
	ctx.Stack.Push(result)
}

func opAddMod(ctx *ExecutionCtx) {
	op1, op2, op3 := ctx.Stack.Pop(), ctx.Stack.Pop(), ctx.Stack.Pop()
	result := uint256.NewInt(0)
	result.AddMod(op1, op2, op3)
	ctx.Stack.Push(result)
}

func opMulMod(ctx *ExecutionCtx) {
	op1, op2, op3 := ctx.Stack.Pop(), ctx.Stack.Pop(), ctx.Stack.Pop()
	result := uint256.NewInt(0)
	result.MulMod(op1, op2, op3)
	ctx.Stack.Push(result)
}

func opReturn(ctx *ExecutionCtx) {
	op1, op2 := ctx.Stack.Pop(), ctx.Stack.Pop()
	ctx.SetReturnData(op1.Uint64(), op2.Uint64())
}

func opJump(ctx *ExecutionCtx) {
	pc := ctx.Stack.Pop().Uint64()
	fmt.Printf("valid jump dests: %v\n", ctx.Jumpdests)
	if _, ok := ctx.Jumpdests[pc]; !ok {
		panic(fmt.Errorf("invalid jump destination %d", pc))
	}
	ctx.SetProgramCounter(pc)
}

func opJumpi(ctx *ExecutionCtx) {
	pc, cond := ctx.Stack.Pop().Uint64(), ctx.Stack.Pop()
	if cond.Cmp(uint256.NewInt(0)) != 0 {
		if _, ok := ctx.Jumpdests[pc]; !ok {
			panic(fmt.Errorf("invalid jump destination %d", pc))
		}
		ctx.SetProgramCounter(pc)
	}
}

func opMload(ctx *ExecutionCtx) {
	offset := ctx.Stack.Pop()
	ctx.Stack.Push(ctx.Memory.LoadWord(offset.Uint64()))
}

func opMstore8(ctx *ExecutionCtx) {
	offset, value := ctx.Stack.Pop(), ctx.Stack.Pop()
	// MSTORE8 pops an offset and a word from the stack,
	// and stores the lowest byte of that word in memory
	value.Mod(value, uint256.NewInt(256))
	ctx.Memory.StoreByte(offset.Uint64(), uint8(value.Uint64()))
}

func opMstore(ctx *ExecutionCtx) {
	offset, value := ctx.Stack.Pop(), ctx.Stack.Pop()
	ctx.Memory.StoreWord(offset.Uint64(), *value)
}

func opSload(ctx *ExecutionCtx) {
	slot := ctx.Stack.Pop()
	value := ctx.Storage.Get(*slot)
	ctx.Stack.Push(value)
}

func opSstore(ctx *ExecutionCtx) {
	slot, value := ctx.Stack.Pop(), ctx.Stack.Pop()
	ctx.Storage.Put(slot, value)
}

func opProgramCounter(ctx *ExecutionCtx) {
	ctx.Stack.Push(uint256.NewInt(ctx.pc))
}

func opMsize(ctx *ExecutionCtx) {
	ctx.Stack.Push(uint256.NewInt(ctx.Memory.ActiveWords() * 32))
}

func opJumpdest(_ *ExecutionCtx) {}

func opDup1(ctx *ExecutionCtx) {
	ctx.Stack.Push(ctx.Stack.Peek(0))
}

func opDup2(ctx *ExecutionCtx) {
	ctx.Stack.Push(ctx.Stack.Peek(1))
}

func opDup3(ctx *ExecutionCtx) {
	ctx.Stack.Push(ctx.Stack.Peek(2))
}

func OpSwap1(ctx *ExecutionCtx) {
	ctx.Stack.Swap(1)
}

func opCalldataLoad(ctx *ExecutionCtx) {
	// geth limits the size of calldata to uint64
	// https://github.com/ethereum/go-ethereum/blob/440c9fcf75d9d5383b72646a65d5e21fa7ab6a26/core/vm/instructions.go
	if offset, overflow := ctx.Stack.Pop().Uint64WithOverflow(); !overflow {
		ctx.Stack.Push(ctx.Calldata.ReadWord(offset))
	}
}

func opCalldataSize(ctx *ExecutionCtx) {
	ctx.Stack.Push(uint256.NewInt(ctx.Calldata.Size()))
}

func opGas(ctx *ExecutionCtx) {
	ctx.Stack.Push(uint256.NewInt(ctx.Gas))
}

func opCodeSize(ctx *ExecutionCtx) {
	ctx.Stack.Push(uint256.NewInt(uint64(len(ctx.code))))
}
