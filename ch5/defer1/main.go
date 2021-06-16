package main

import "fmt"

func main() {
	f(3)
}

func f(x int) {
	if x > 0 {
		fmt.Printf("f(%d)\n", x)
		defer fmt.Printf("defer %d\n", x)
		f(x - 1)
	}
}
