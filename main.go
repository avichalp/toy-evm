package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/avichalp/toy-evm/evm"
)

func main() {
	var (
		code     string
		calldata string
		gas      uint64
	)
	flag.StringVar(&code, "code", "0x0", "hex data of the code to run")
	flag.StringVar(&calldata, "calldata", "0x0", "hex data to use as input")
	flag.Uint64Var(&gas, "gas", 5, "number of steps the VM will execute")
	flag.Parse()
	fmt.Printf("code: %s, calldata %s, gas %d\n", code, calldata, gas)

	evm.Init()
	fmt.Printf("\n")

	ctx, cancel := context.WithCancel(context.Background())
	ectx := evm.NewExecutionCtx(
		ctx,
		cancel,
		evm.HexToBytes(code),
		evm.NewStack(),
		evm.NewMemory(),
		gas,
	)
	evm.Run(ectx)
}
