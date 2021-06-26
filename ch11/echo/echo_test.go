package echo_test

import (
	. "../echo"
	"bytes"
	"fmt"
	"testing"
)

func TestEcho(t *testing.T) {
	var tests = []struct {
		newline bool
		sep     string
		args    []string
		want    string
	}{
		{true, "", []string{}, "\n"},
		{false, "", []string{}, ""},
		{true, "\t", []string{"one", "two", "three"}, "one\ttwo\tthree\n"},
		{true, ",", []string{"a", "b", "c"}, "a,b,c\n"},
		{false, ":", []string{"1", "2", "3"}, "1:2:3"},
	}

	for _, test := range tests {
		descr := fmt.Sprintf("echo(%v, %q, %q)",
			test.newline, test.sep, test.args)

		Out = new(bytes.Buffer) // captured output

		t.Run(descr, func(t *testing.T) {
			if err := Echo(test.newline, test.sep, test.args); err != nil {
				t.Errorf("%s failed: %v", descr, err)
			}
			got := Out.(*bytes.Buffer).String()
			if got != test.want {
				t.Errorf("%s = %q, want %q", descr, got, test.want)
			}
		})
	}
}
