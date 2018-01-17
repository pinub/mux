package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pinub/mux"
)

func index() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome!\n")
	})
}

func hello() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello\n")
	})
}

func main() {
	m := mux.New()
	m.Get("/", index())
	m.Get("/hello", hello())

	log.Fatal(http.ListenAndServe(":8080", m))
}
