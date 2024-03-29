# mux [![GitHub Actions](https://github.com/pinub/mux/workflows/CI/badge.svg)](https://github.com/pinub/mux/actions) [![GoDoc](https://godoc.org/github.com/pinub/mux?status.svg)](https://godoc.org/github.com/pinub/mux)

mux is a high performance HTTP request router, also called multiplexer or just _mux_.

## Example

~~~go
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
~~~

The Muxer matches incoming requests by the method and path and delegates to that assiciated function. Currently GET, POST, PUT, PATCH, DELETE and OPTIONS are supported methods.

Named parameters are not supported.

~~~
Path: /foo/bar

Requests:

    /foo/bar        matches the function
    /foo/bar/       doesn't match, but redirects to /foo/bar
    /foo/foo        doesn't match
    /foo            doesn't match
~~~

## Acknowledge

Parts of the source are copied from Julien Schmidt famous [httprouter](https://github.com/julienschmidt/httprouter). So parts of the code are `Copyright (c) 2013 Julien Schmidt. All rights reserved.`
