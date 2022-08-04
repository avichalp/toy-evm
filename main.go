package main

import (
	"context"
	"flag"
	"fmt"
)

func decodeOpcode(ctx context.Context) context.Context {
	fmt.Println("decoding opcode")
	return ctx
}

func execute(numSteps *int, cancel context.CancelFunc) {
	// todo: if number of instruction exceeded or excecution is complete
	if *numSteps >= 3 {
		fmt.Println("NUM OF STEPS REACHED")
		cancel()
		return
	}
	fmt.Printf("exectuing opcode\n")
	*numSteps++
}

func run(code, calldata string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	numSteps := 0

	for {
		select {
		case <-ctx.Done():
			fmt.Println("NUM STEPS EXECUTED", numSteps)
			cancel()
			return
		default:
			decodeOpcode(ctx)
			execute(&numSteps, cancel)
		}
	}
}

func main() {
	var code string
	var calldata string
	flag.StringVar(&code, "code", "0x0", "hex data of the code to run")
	flag.StringVar(&calldata, "calldata", "0x0", "hex data to use as input")
	flag.Parse()
	fmt.Printf("code: %s, calldata %s\n", code, calldata)
	run(code, calldata)

}
