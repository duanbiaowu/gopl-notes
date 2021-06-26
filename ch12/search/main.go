package main

import (
	"../params"
	"fmt"
	"log"
	"net/http"
)

// search implements the /search URL endpoint.
func search(resp http.ResponseWriter, req *http.Request)  {
	var data struct {
		Labels    []string `http:"l"`
		MaxResult int      `http:"max"`
		Exact     bool     `http:"x"`
	}
	data.MaxResult = 10 // set default
	if err := params.Unpack(req, &data); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest) // 400
		return
	}

	// ...rest of handler...
	_, _ = fmt.Fprintf(resp, "Search: %+v\n", data)
}

func main() {
	http.HandleFunc("/search", search)
	log.Fatal(http.ListenAndServe("localhost:8081", nil))
}
