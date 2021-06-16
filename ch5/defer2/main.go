package main

import (
	"fmt"
	"os"
	"runtime"
)

func f(x int) {
	if x > 0 {
		fmt.Printf("f(%d)\n", x)
		defer fmt.Printf("defer %d\n", x)
		f(x - 1)
	}
}

func printStack() {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	_, _ = os.Stdout.Write(buf[:n])
}

func main() {
	defer printStack()
	f(3)
}
