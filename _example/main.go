package main

import (
	"log"
	"net/http"

	"github.com/pinub/mux/v3"
)

func index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`Welcome!`))
	}
}

func hello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`Welcome!`))
	}
}

func main() {
	m := mux.New()
	m.Get("/", index())
	m.Get("/hello", hello())

	log.Fatal(http.ListenAndServe(":8080", m))
}
