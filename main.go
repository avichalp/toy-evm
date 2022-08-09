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
		0x53,
		"MSTORE8",
		func(ctx *ExecutionCtx) {
			op1, op2 := ctx.stack.Pop(), ctx.stack.Pop()
			// MSTORE8 pops an offset and a word from the stack,
			// and stores the lowest byte of that word in memory
			op2.Mod(op2, uint256.NewInt(256))
			ctx.memory.Store(op1.Uint64(), uint8(op2.Uint64()))
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
			ctx.SetProgramCounter(pc)
		},
	)

	fmt.Printf("\n")

	run(code, calldata)
}
