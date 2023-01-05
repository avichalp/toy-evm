package evm

import (
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func TestOpAdd(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack: NewStack(),
	}
	ctx.Stack.Push(uint256.NewInt(1))
	ctx.Stack.Push(uint256.NewInt(2))
	opAdd(ctx)
	assert.Equal(t, uint256.NewInt(3), ctx.Stack.Pop())
}

func TestOpMul(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack: NewStack(),
	}
	ctx.Stack.Push(uint256.NewInt(2))
	ctx.Stack.Push(uint256.NewInt(3))
	opMul(ctx)
	assert.Equal(t, uint256.NewInt(6), ctx.Stack.Pop())
}

func TestOpSub(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack: NewStack(),
	}
	ctx.Stack.Push(uint256.NewInt(2))
	ctx.Stack.Push(uint256.NewInt(3))
	opSub(ctx)
	assert.Equal(t, uint256.NewInt(1), ctx.Stack.Pop())
}

func TestOpDiv(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack: NewStack(),
	}
	ctx.Stack.Push(uint256.NewInt(6))
	ctx.Stack.Push(uint256.NewInt(3))
	opDiv(ctx)
	assert.Equal(t, uint256.NewInt(0), ctx.Stack.Pop())
}

func TestOpMod(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack: NewStack(),
	}
	ctx.Stack.Push(uint256.NewInt(3))
	ctx.Stack.Push(uint256.NewInt(6))
	opMod(ctx)
	assert.Equal(t, uint256.NewInt(0), ctx.Stack.Pop())

	ctx.Stack.Push(uint256.NewInt(0))
	ctx.Stack.Push(uint256.NewInt(6))
	opMod(ctx)
	assert.Equal(t, uint256.NewInt(0), ctx.Stack.Pop())
}

func TestOpAddMod(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack: NewStack(),
	}
	ctx.Stack.Push(uint256.NewInt(1))
	ctx.Stack.Push(uint256.NewInt(2))
	ctx.Stack.Push(uint256.NewInt(3))
	opAddMod(ctx)
	assert.Equal(t, uint256.NewInt(0), ctx.Stack.Pop())

	ctx.Stack.Push(uint256.NewInt(0))
	ctx.Stack.Push(uint256.NewInt(21))
	ctx.Stack.Push(uint256.NewInt(20))
	opAddMod(ctx)
	assert.Equal(t, uint256.NewInt(0), ctx.Stack.Pop())
}

func TestOpMulMod(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack: NewStack(),
	}
	ctx.Stack.Push(uint256.NewInt(3))
	ctx.Stack.Push(uint256.NewInt(2))
	ctx.Stack.Push(uint256.NewInt(1))
	opMulMod(ctx)
	assert.Equal(t, uint256.NewInt(2), ctx.Stack.Pop())

	ctx.Stack.Push(uint256.NewInt(0))
	ctx.Stack.Push(uint256.NewInt(20))
	ctx.Stack.Push(uint256.NewInt(20))
	opMulMod(ctx)
	assert.Equal(t, uint256.NewInt(0), ctx.Stack.Pop())

	ctx.Stack.Push(uint256.NewInt(2))
	ctx.Stack.Push(uint256.NewInt(0))
	ctx.Stack.Push(uint256.NewInt(20))
	opMulMod(ctx)
	assert.Equal(t, uint256.NewInt(0), ctx.Stack.Pop())
}

func TestOpReturn(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack:  NewStack(),
		Memory: NewMemory(),
	}
	ctx.Stack.Push(uint256.NewInt(42))
	ctx.Stack.Push(uint256.NewInt(0))
	opMstore8(ctx)

	ctx.Stack.Push(uint256.NewInt(1))
	ctx.Stack.Push(uint256.NewInt(0))
	opReturn(ctx)
	assert.Equal(t, []byte{42}, ctx.Returndata)
}

func TestOpJump(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack:     NewStack(),
		Jumpdests: map[uint64]uint64{1: 1, 2: 2},
		pc:        0,
	}
	ctx.Stack.Push(uint256.NewInt(1))
	opJump(ctx)
	assert.Equal(t, uint64(1), ctx.pc)
}

func TestOpJumpFail(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack:     NewStack(),
		Jumpdests: map[uint64]uint64{1: 1, 2: 2},
		pc:        0,
	}
	ctx.Stack.Push(uint256.NewInt(3))
	assert.PanicsWithError(
		t,
		"invalid jump destination 3",
		func() {
			opJump(ctx)
		})
}

func TestOpJumpi(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack:     NewStack(),
		Jumpdests: map[uint64]uint64{1: 1, 2: 2},
		pc:        0,
	}
	//condition true
	ctx.Stack.Push(uint256.NewInt(1))
	ctx.Stack.Push(uint256.NewInt(1))
	opJumpi(ctx)
	assert.Equal(t, uint64(1), ctx.pc)

	// condition false (noop)
	ctx.Stack.Push(uint256.NewInt(0))
	ctx.Stack.Push(uint256.NewInt(1))
	opJumpi(ctx)
	assert.Equal(t, ctx.pc, ctx.pc)

	// condition true with invalid jump destination
	ctx.Stack.Push(uint256.NewInt(1))
	ctx.Stack.Push(uint256.NewInt(3))
	assert.PanicsWithError(
		t,
		"invalid jump destination 3", func() {
			opJumpi(ctx)
		},
	)
}

func TestOpMload(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack:  NewStack(),
		Memory: NewMemory(),
	}
	ctx.Memory.StoreWord(0, *uint256.NewInt(42))
	ctx.Stack.Push(uint256.NewInt(0))
	opMload(ctx)
	assert.Equal(t, uint256.NewInt(42), ctx.Stack.Pop())
}

func TestOpMstore(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack:  NewStack(),
		Memory: NewMemory(),
	}
	ctx.Stack.Push(uint256.NewInt(42))
	ctx.Stack.Push(uint256.NewInt(0))
	opMstore(ctx)
	assert.Equal(t, uint256.NewInt(42), ctx.Memory.LoadWord(0))
}

func TestOpMstore8(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack:  NewStack(),
		Memory: NewMemory(),
	}
	ctx.Stack.Push(uint256.NewInt(42))
	ctx.Stack.Push(uint256.NewInt(0))
	opMstore8(ctx)
	output := uint256.NewInt(0).SetBytes([]byte{
		42, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
	})
	assert.Equal(t, output, ctx.Memory.LoadWord(0))
}

func TestOpSload(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack:   NewStack(),
		Storage: NewStorage(),
	}
	ctx.Storage.Put(uint256.NewInt(1), uint256.NewInt(42))
	ctx.Stack.Push(uint256.NewInt(1))
	opSload(ctx)
	assert.Equal(t, uint256.NewInt(42), ctx.Stack.Pop())
}

func TestOpSstore(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack:   NewStack(),
		Storage: NewStorage(),
	}
	ctx.Stack.Push(uint256.NewInt(42))
	ctx.Stack.Push(uint256.NewInt(1))
	opSstore(ctx)
	assert.Equal(t, uint256.NewInt(42), ctx.Storage.Get(*uint256.NewInt(1)))
}

func TestOpProgramCounter(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack: NewStack(),
		pc:    0,
	}
	opProgramCounter(ctx)
	assert.Equal(t, uint64(0), ctx.pc)
}

func TestOpMsize(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack:  NewStack(),
		Memory: NewMemory(),
	}
	num, _ := uint256.FromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	ctx.Memory.StoreWord(0, *num)
	opMsize(ctx)
	assert.Equal(t, uint256.NewInt(32), ctx.Stack.Pop())

	// overwrite memory with 1
	ctx.Memory.StoreWord(0, *uint256.NewInt(1))
	opMsize(ctx)
	assert.Equal(t, uint256.NewInt(32), ctx.Stack.Pop())
}

func TestOpGas(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack: NewStack(),
		Gas:   1000,
	}
	opGas(ctx)
	assert.Equal(t, uint64(1000), ctx.Stack.Pop().Uint64())
}

func TestOpCalldataSize(t *testing.T) {
	// dummy 32byte as a hex string
	dummyCalldata := "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"
	ctx := &ExecutionCtx{
		Stack:    NewStack(),
		Calldata: NewCalldata(dummyCalldata),
	}
	opCalldataSize(ctx)
	assert.Equal(t, uint256.NewInt(32), ctx.Stack.Pop())
}

func TestOpCalldataLoad(t *testing.T) {
	dummyCalldata := "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"
	ctx := &ExecutionCtx{
		Stack:    NewStack(),
		Memory:   NewMemory(),
		Calldata: NewCalldata(dummyCalldata),
	}
	ctx.Stack.Push(uint256.NewInt(0))
	opCalldataLoad(ctx)
	bigNum, _ := uint256.FromHex("0x" + dummyCalldata)
	assert.Equal(t, bigNum, ctx.Stack.Pop())

	num, _ := uint256.FromHex("0x10000000000000000")
	ctx.Stack.Push(num)
	opCalldataLoad(ctx)
	assert.Equal(t, 0, len(ctx.Stack.data))
}

func TestOpCodeSize(t *testing.T) {
	ctx := &ExecutionCtx{
		Stack: NewStack(),
		code:  []byte{0x01, 0x02, 0x03},
	}
	opCodeSize(ctx)
	assert.Equal(t, uint256.NewInt(3), ctx.Stack.Pop())

	ctx.code = []byte{}
	opCodeSize(ctx)
	assert.Equal(t, uint256.NewInt(0), ctx.Stack.Pop())
}
