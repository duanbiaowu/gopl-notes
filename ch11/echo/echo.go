package echo

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	n = flag.Bool("n", false, "omit trailing newline")
	s = flag.String("s", " ", "separator")
)

var Out io.Writer = os.Stdout // modified during testing

func main() {
	flag.Parse()
	if err := Echo(!*n, *s, flag.Args()); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "echo: %v\n", err)
		os.Exit(1)
	}
}

func Echo(newline bool, sep string, args []string) error {
	_, _ = fmt.Fprintf(Out, strings.Join(args, sep))
	if newline {
		_, _ = fmt.Fprintln(Out)
	}
	return nil
}
