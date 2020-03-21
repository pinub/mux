// Package mux is a high performance HTTP request router, also called
// multiplexer or just mux.
//
// Example:
//
//  package main
//
//  import (
//  	"log"
//  	"net/http"
//
//  	"github.com/pinub/mux/v3"
//  )
//
//  func index(w http.ResponseWriter, r *http.Request) {
//  	w.Write([]byte(`Welcome to index!`))
//  }
//
//  func hello(w http.ResponseWriter, r *http.Request) {
//  	w.Write([]byte(`Welcome to hello!`))
//  }
//
//  func main() {
//  	m := mux.New()
//  	m.Get("/", index)
//  	m.Get("/hello", hello)
//
//  	log.Fatal(http.ListenAndServe(":8080", m))
//  }
//
// The Muxer matches incoming requests by the method and path and delegates
// to that assiciated handler.
// Currently GET, POST, PUT, PATCH, DELETE and OPTIONS are supported methods.
//
// Named parameters are not supported.
//
// Path: /foo/bar
//
// Requests:
//  /foo/bar        matches the handler
//  /foo/bar/       doesn't match, but redirects to /foo/bar
//  /foo/foo        doesn't match
//  /foo            doesn't match
package mux

import (
	"net/http"
	"strings"
)

// Router is a http.Handler used to dispatch request to different handlers.
type Router struct {
	routes map[string]http.HandlerFunc

	// Enables automatic redirection if the requested path doesn't match but
	// a handler with a path without the trailing slash exists. Default: true
	RedirectTrailingSlash bool

	// Enables 'Method Not Allowed' responses when a handler for the the
	// path, but not the requested method exists. Default: true
	HandleMethodNotAllowed bool

	// Custom http.Handler which is called when no handler was found for the
	// requested route. Defaults: http.NotFound.
	NotFound http.HandlerFunc
}

// New initializes the Router.
// All configurable options are enabled by default.
func New() *Router {
	return &Router{
		RedirectTrailingSlash:  true,
		HandleMethodNotAllowed: true,
	}
}

// Make sure Router conforms with http.Handler interface.
var _ http.Handler = New()

// Head registers a new request handle for the HEAD method and the given path.
func (r *Router) Head(path string, h http.HandlerFunc) {
	r.add(http.MethodHead, path, h)
}

// Get registers a new request handle for the GET and HEAD method and the given path.
func (r *Router) Get(path string, h http.HandlerFunc) {
	r.add(http.MethodHead, path, h)
	r.add(http.MethodGet, path, h)
}

// Post registers a new request handle for the POST method and the given path.
func (r *Router) Post(path string, h http.HandlerFunc) {
	r.add(http.MethodPost, path, h)
}

// Put registers a new request handle for the PUT method and the given path.
func (r *Router) Put(path string, h http.HandlerFunc) {
	r.add(http.MethodPut, path, h)
}

// Delete registers a new request handle for the DELETE method and the given path.
func (r *Router) Delete(path string, h http.HandlerFunc) {
	r.add(http.MethodDelete, path, h)
}

// Options registers a new request handle for the OPTIONS method and the given path.
func (r *Router) Options(path string, h http.HandlerFunc) {
	r.add(http.MethodOptions, path, h)
}

// Patch registers a new request handle for the PATCH method and the given path.
func (r *Router) Patch(path string, h http.HandlerFunc) {
	r.add(http.MethodPatch, path, h)
}

func (r *Router) add(method string, path string, h http.HandlerFunc) {
	if path[0] != '/' {
		panic("Path must begin with '/' in path '" + path + "'")
	}

	if r.routes == nil {
		r.routes = make(map[string]http.HandlerFunc)
	}
	r.routes[method+path] = h

	// redirect for paths ending with a '/'
	if r.RedirectTrailingSlash {
		n := len(path)
		if n > 1 && path[n-1] != '/' {
			r.routes[method+path+"/"] = slashRedirect
		}
	}
}

func slashRedirect(w http.ResponseWriter, r *http.Request) {
	u := *r.URL
	u.Path = u.Path[:len(u.Path)-1]

	code := http.StatusMovedPermanently
	if r.Method != http.MethodGet {
		code = http.StatusPermanentRedirect
	}

	http.Redirect(w, r, u.String(), code)
}

// ServeHTTP makes this router implement the http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method

	// fix for forms with given _method value
	if method == http.MethodPost {
		if formMethod := req.FormValue("_method"); formMethod != "" {
			method = strings.ToUpper(formMethod)
		}
	}

	if h, ok := r.routes[method+path]; ok {
		h.ServeHTTP(w, req)
		return
	}

	// method not allowed
	if r.HandleMethodNotAllowed {
		if allowed := r.allowed(path); len(allowed) > 0 {
			w.Header().Set("Allow", allowed)
			if method != http.MethodOptions {
				http.Error(w,
					http.StatusText(http.StatusMethodNotAllowed),
					http.StatusMethodNotAllowed,
				)
			}

			return
		}
	}

	if r.NotFound != nil {
		r.NotFound.ServeHTTP(w, req)
	} else {
		http.NotFound(w, req)
	}
}

func (r *Router) allowed(path string) (allowed string) {
	methods := []string{
		http.MethodHead,
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodOptions,
	}

	for _, method := range methods {
		if _, ok := r.routes[method+path]; ok {
			if len(allowed) == 0 {
				allowed = method
			} else {
				allowed += ", " + method
			}
		}
	}

	return
}
