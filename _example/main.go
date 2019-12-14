package main

import (
	"log"
	"net/http"
	"time"

	"github.com/pinub/mux/v3"
)

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`Welcome to index!`))
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`Welcome to hello!`))
}

// a middleware
func logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := time.Now()
		next.ServeHTTP(w, r)

		log.Printf("%s %s took %s", r.Method, r.URL.String(), time.Since(s))
	}
}

func main() {
	m := mux.New()
	m.Get("/", logging(index))
	m.Get("/hello", logging(hello))

	log.Fatal(http.ListenAndServe(":8080", m))
}
