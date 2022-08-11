package main

import (
	"flag"
	"fmt"

	"github.com/avichalp/toy-evm/evm"
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
	evm.Run(code, calldata, steps)
}
