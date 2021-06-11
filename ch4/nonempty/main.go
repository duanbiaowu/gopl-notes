package main

import "fmt"

func nonempty(strings []string) []string {
	i := 0
	for _, s := range strings {
		if s != "" {
			strings[i] = s
			i++
		}
	}
	return strings[:i]
}

func nonempty2(strings []string) []string {
	out := strings[:0]
	for _, s := range strings {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func main() {
	f := nonempty
	f2 := nonempty2
	data := []string{"one", "", "three"}
	data2 := []string{"hello", "", "world"}
	fmt.Printf("%q\n", f(data))
	fmt.Printf("%q\n", f2(data2))
	fmt.Printf("%q\n", data)
	fmt.Printf("%q\n", data2)
}
