package main

import (
	"flag"
	"fmt"
	"math"

	"github.com/holiman/uint256"
)

var Instructions = make([]*Instruction, 0)
var InstructionByOpcode = make(map[byte]*Instruction)

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

func main() {
	var code string
	var calldata string
	flag.StringVar(&code, "code", "0x0", "hex data of the code to run")
	flag.StringVar(&calldata, "calldata", "0x0", "hex data to use as input")
	flag.Parse()
	fmt.Printf("code: %s, calldata %s\n", code, calldata)

	RegisterInstruction(0x0, "STOP", func(ctx *ExecutionCtx) { ctx.Stop() })
	RegisterInstruction(
		0x60,
		"PUSH1",
		func(ctx *ExecutionCtx) {
			ctx.stack.Push(uint256.NewInt(uint64(ctx.ReadCode(1))))
		})
	RegisterInstruction(
		0x01,
		"ADD",
		func(ctx *ExecutionCtx) {
			op1, op2 := ctx.stack.Pop(), ctx.stack.Pop()
			result := uint256.NewInt(0)
			mod := uint256.NewInt(uint64(math.Pow(2, 256)))
			result.AddMod(op1, op2, mod)
			ctx.stack.Push(result)
		})
	RegisterInstruction(
		0x02,
		"MUL",
		func(ctx *ExecutionCtx) {
			op1, op2 := ctx.stack.Pop(), ctx.stack.Pop()
			result := uint256.NewInt(0)
			mod := uint256.NewInt(uint64(math.Pow(2, 256)))
			result.MulMod(op1, op2, mod)
			ctx.stack.Push(result)
		})
	RegisterInstruction(
		0x03,
		"SUB",
		func(ctx *ExecutionCtx) {
			op1, op2 := ctx.stack.Pop(), ctx.stack.Pop()
			result := uint256.NewInt(0)
			mod := uint256.NewInt(uint64(math.Pow(2, 256)))
			result.Sub(op1, op2)
			result.Mod(result, mod)
			ctx.stack.Push(result)
		},
	)
	RegisterInstruction(
		0xF3,
		"RETURN",
		func(ctx *ExecutionCtx) {
			op1, op2 := ctx.stack.Pop(), ctx.stack.Pop()
			// TODO: these arguments should be uint256
			ctx.SetReturnData(op1.Uint64(), op2.Uint64())
		},
	)
	RegisterInstruction(
		0x56,
		"JUMP",
		func(ctx *ExecutionCtx) {
			pc := ctx.stack.Pop().Uint64()
			fmt.Printf("valid jump dests: %v\n", ctx.jumpdests)
			if _, ok := ctx.jumpdests[pc]; !ok {
				panic(fmt.Errorf("invalid jump destination %d", pc))
			}
			ctx.SetProgramCounter(pc)
		},
	)
	RegisterInstruction(
		0x57,
		"JUMPI",
		func(ctx *ExecutionCtx) {
			pc, cond := ctx.stack.Pop().Uint64(), ctx.stack.Pop()
			if cond.Cmp(uint256.NewInt(0)) != 0 {
				if _, ok := ctx.jumpdests[pc]; !ok {
					panic(fmt.Errorf("invalid jump destination %d", pc))
				}
				ctx.SetProgramCounter(pc)
			}
		},
	)
	RegisterInstruction(
		0x51,
		"MLOAD",
		func(ctx *ExecutionCtx) {
			offset := ctx.stack.Pop()
			ctx.stack.Push(ctx.memory.LoadWord(offset.Uint64()))
		},
	)
	RegisterInstruction(
		0x53,
		"MSTORE8",
		func(ctx *ExecutionCtx) {
			op1, op2 := ctx.stack.Pop(), ctx.stack.Pop()
			// MSTORE8 pops an offset and a word from the stack,
			// and stores the lowest byte of that word in memory
			op2.Mod(op2, uint256.NewInt(256))
			ctx.memory.StoreByte(op1.Uint64(), uint8(op2.Uint64()))
		},
	)
	RegisterInstruction(
		0x52,
		"MSTORE",
		func(ctx *ExecutionCtx) {
			offset, value := ctx.stack.Pop(), ctx.stack.Pop()
			ctx.memory.StoreWord(offset.Uint64(), *value)
		},
	)
	RegisterInstruction(
		0x58,
		"PC",
		func(ctx *ExecutionCtx) { ctx.stack.Push(uint256.NewInt(ctx.pc)) },
	)
	RegisterInstruction(
		0x59,
		"MSIZE",
		func(ctx *ExecutionCtx) {
			ctx.stack.Push(uint256.NewInt(ctx.memory.ActiveWords() * 32))
		},
	)
	RegisterInstruction(
		0x5B,
		"JUMPDEST",
		func(_ *ExecutionCtx) {},
	)
	RegisterInstruction(
		0x80,
		"DUP1",
		func(ctx *ExecutionCtx) {
			ctx.stack.Push(ctx.stack.Peek(0))
		},
	)
	RegisterInstruction(
		0x81,
		"DUP2",
		func(ctx *ExecutionCtx) {
			ctx.stack.Push(ctx.stack.Peek(1))
		},
	)
	RegisterInstruction(
		0x82,
		"DUP3",
		func(ctx *ExecutionCtx) {
			ctx.stack.Push(ctx.stack.Peek(2))
		},
	)
	RegisterInstruction(
		0x90,
		"SWAP1",
		func(ctx *ExecutionCtx) {
			ctx.stack.Swap(1)
		},
	)
	RegisterInstruction(
		0x35,
		"CALLDATALOAD",
		func(ctx *ExecutionCtx) {
			// todo: theoritially there is not limit to calldata
			// is uint64 safe size for calldata's byte array?
			// geth limits the size of calldata to uint64
			// https://github.com/ethereum/go-ethereum/blob/440c9fcf75d9d5383b72646a65d5e21fa7ab6a26/core/vm/instructions.go
			if offset, overflow := ctx.stack.Pop().Uint64WithOverflow(); !overflow {
				ctx.stack.Push(ctx.calldata.ReadWord(offset))
			}
		},
	)
	fmt.Printf("\n")

	run(code, calldata)
}
