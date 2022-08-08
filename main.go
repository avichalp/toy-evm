package main

import (
	"flag"
	"fmt"
	"math"

	"github.com/holiman/uint256"
)

var Instructions = make([]*Instruction, 0)
var InstructionByOpcode = make(map[uint256.Int]*Instruction)

func RegisterInstruction(opcode *uint256.Int, name string, executeFn ExecuteFn) *Instruction {
	inst := &Instruction{
		opcode:    opcode,
		name:      name,
		executeFn: executeFn,
	}
	Instructions = append(Instructions, inst)

	InstructionByOpcode[*opcode] = inst
	fmt.Printf("registering %s: %v\n", opcode, inst)
	return inst
}

func main() {
	var code string
	var calldata string
	flag.StringVar(&code, "code", "0x0", "hex data of the code to run")
	flag.StringVar(&calldata, "calldata", "0x0", "hex data to use as input")
	flag.Parse()
	fmt.Printf("code: %s, calldata %s\n", code, calldata)

	RegisterInstruction(uint256.NewInt(0), "STOP", func(ctx *ExecutionCtx) { ctx.Stop() })

	if pushOP, err := uint256.FromHex("0x60"); err == nil {
		RegisterInstruction(
			pushOP,
			"PUSH1",
			func(ctx *ExecutionCtx) {
				ctx.stack.Push(ctx.ReadCode(1))
			})
	}

	RegisterInstruction(
		uint256.NewInt(1),
		"ADD",
		func(ctx *ExecutionCtx) {
			op1, op2 := ctx.stack.Pop(), ctx.stack.Pop()
			result := uint256.NewInt(0)
			mod := uint256.NewInt(uint64(math.Pow(2, 256)))
			result.AddMod(op1, op2, mod)
			ctx.stack.Push(result)
		})

	RegisterInstruction(
		uint256.NewInt(2),
		"MUL",
		func(ctx *ExecutionCtx) {
			op1, op2 := ctx.stack.Pop(), ctx.stack.Pop()
			result := uint256.NewInt(0)
			mod := uint256.NewInt(uint64(math.Pow(2, 256)))
			result.MulMod(op1, op2, mod)
			ctx.stack.Push(result)
		})

	MSTORE8OP, err := uint256.FromHex("0x53")
	if err != nil {
		panic("Cannot register MSTORE8 instruction")
	}
	RegisterInstruction(
		MSTORE8OP,
		"MSTORE8",
		func(ctx *ExecutionCtx) {
			op1, op2 := ctx.stack.Pop(), ctx.stack.Pop()
			// MSTORE8 pops an offset and a word from the stack,
			// and stores the lowest byte of that word in memory
			op2.Mod(op2, uint256.NewInt(256))
			ctx.memory.Store(op1.Uint64(), uint8(op2.Uint64()))
		},
	)

	RETURNOP, err := uint256.FromHex("0xf3")
	if err != nil {
		panic("Cannot register RETURN instruction")
	}
	RegisterInstruction(
		RETURNOP,
		"RETURN",
		func(ctx *ExecutionCtx) {
			op1, op2 := ctx.stack.Pop(), ctx.stack.Pop()
			// TODO: these arguments should be uint256
			ctx.SetReturnData(op1.Uint64(), op2.Uint64())
		},
	)

	fmt.Printf("\n")

	run(code, calldata)
}
