package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func xmlselect(r io.Reader) {
	doc := xml.NewDecoder(r)
	var stack []string
	var checkStack = []string{"div", "div", "h2"}
	for {
		tok, err := doc.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "xmlselect: %v\n", err)
			os.Exit(1)
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			stack = append(stack, tok.Name.Local)
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if containsAll(stack, checkStack) {
				fmt.Printf("%s: %s\n", strings.Join(stack, " "), tok)
			}
		}
	}
}

func containsAll(x, y []string) bool {
	for len(y) <= len(x) {
		if len(y) == 0 {
			return true
		}
		if x[0] == y[0] {
			y = y[1:]
		}
		x = x[1:]
	}
	return false
}

func fetch(w io.Writer) {
	url := "http://www.w3.org/TR/2006/REC-xml11-20060816"
	resp, err := http.Get(url)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "fetch : %v\n", err)
		os.Exit(1)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "fetch reading %s: %v\n", url, err)
		os.Exit(1)
	}
	_, _ = fmt.Fprintf(w, "%s\n", b)
}

func main() {
	buf := bytes.NewBuffer(nil)
	fetch(buf)
	xmlselect(buf)
}
