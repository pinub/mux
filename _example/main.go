package main

import (
	"log"
	"net/http"

	"github.com/pinub/mux/v3"
)

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`Welcome to index!`))
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`Welcome to hello!`))
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
        log.Printf("%s", "Hello from Middleware")
		next.ServeHTTP(rw, r)
	})
}

func main() {
	m := mux.New()
	m.Get("/", index)
	m.Get("/hello", hello)

	m.Use(middleware)

	log.Fatal(http.ListenAndServe(":8080", m))
}
