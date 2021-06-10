package main

import (
	"bytes"
	"fmt"
)

func intToStrings(values []int) string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, v := range values {
		if i > 0 {
			buf.WriteString(", ")
		}
		_, _ = fmt.Fprintf(&buf, "%d", v)
	}
	buf.WriteByte(']')
	return buf.String()
}

func main() {
	fmt.Println(intToStrings([]int{1, 2, 3}))
	fmt.Println([]int{1, 2, 3})
}
