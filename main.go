package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/avichalp/toy-evm/evm"
	"github.com/status-im/keycard-go/hexutils"
)

func main() {
	var (
		code     string
		calldata string
		steps    int
	)
	flag.StringVar(&code, "code", "0x0", "hex data of the code to run")
	flag.StringVar(&calldata, "calldata", "0x0", "hex data to use as input")
	flag.IntVar(&steps, "steps", 5, "number of steps the VM will execute")
	flag.Parse()
	fmt.Printf("code: %s, calldata %s, gas %d\n", code, calldata, steps)

	evm.Init()
	fmt.Printf("\n")

	ctx, cancel := context.WithCancel(context.Background())
	ectx := evm.NewExecutionCtx(
		ctx,
		cancel,
		hexutils.HexToBytes(code),
		evm.NewStack(),
		evm.NewMemory(),
		steps,
	)
	evm.Run(ectx)
}
