package main

import (
	"html/template"
	"log"
	"net/http"
)

func main()  {
	const templ = `<p>A: {{.A}}</p><p>B: {{.B}}</p>`
	t := template.Must(template.New("escape").Parse(templ))
	var data struct {
		A string		// untrusted plain text
		B template.HTML	// trusted HTML
	}
	data.A = "<b>Hello!</b>"
	data.B = "<b>Hello!</b>"

	handler := func(w http.ResponseWriter, r *http.Request) {
		if err := t.Execute(w, data); err != nil {
			log.Fatal(err)
		}
	}

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe("localhost:8081", nil))
}
