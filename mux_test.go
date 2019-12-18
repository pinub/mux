package mux

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var h = func(w http.ResponseWriter, r *http.Request) {}

func TestRoutes(t *testing.T) {
	t.Parallel()
	for i, tt := range []struct {
		method string
		path   string
		code   int
	}{
		{http.MethodGet, "/", 200},
		{http.MethodGet, "/blah", 200},
		{http.MethodPost, "/", 200},
	} {
		m := New()
		m.add(tt.method, tt.path, h)

		res := httptest.NewRecorder()
		m.ServeHTTP(res, newRequest(tt.method, tt.path, nil))
		if res.Code != tt.code {
			t.Errorf("[%d]: for path %q: got code %d; want %d", i, tt.path, res.Code, tt.code)
		}
	}
}

func TestNotFound(t *testing.T) {
	m := New()
	m.Get("/foo", h)
	m.Get("/bar", h)

	for _, path := range []string{"/foobar", "/test", "/another/url"} {
		res := httptest.NewRecorder()
		m.ServeHTTP(res, newRequest("GET", path, nil))
		if res.Code != http.StatusNotFound {
			t.Errorf("for path %q: got code %d; want %d", path, res.Code, http.StatusNotFound)
		}
	}
}

func TestNotFoundCustomHandler(t *testing.T) {
	m := New()
	m.NotFound = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(123)
	}
	m.Get("/foo", h)
	res := httptest.NewRecorder()
	m.ServeHTTP(res, newRequest("GET", "/bar", nil))
	if res.Code != 123 {
		t.Errorf("for path %q: got code %d; want %d", "/foo", res.Code, 123)
	}
}

func TestRedirectTrailingSlash(t *testing.T) {
	m := New()

	codes := map[int]bool{
		http.StatusPermanentRedirect: true,
		http.StatusMovedPermanently:  true,
	}

	for i, tt := range []struct {
		method   string
		path     string
		code     int
		addToMux bool
	}{
		{"GET", "/foo", 200, true},
		{"GET", "/foo/", 301, false},
		{"GET", "/bar/", 200, true},
		{"GET", "/blah/", 404, false},
		{"POST", "/foobar", 200, true},
		{"POST", "/foobar/", 308, false},
	} {
		if tt.addToMux {
			m.add(tt.method, tt.path, h)
		}

		res := httptest.NewRecorder()
		m.ServeHTTP(res, newRequest(tt.method, tt.path, nil))
		if res.Code != tt.code {
			t.Errorf("[%d]: for path %q: got code %d; want %d", i, tt.path, res.Code, tt.code)
		}
		location := res.Header().Get("Location")
		if codes[res.Code] && location != strings.TrimRight(tt.path, "/") {
			t.Errorf("[%d]: for path %q: got location %s; want %s", i, tt.path, location, strings.TrimRight(tt.path, "/"))
		}

	}
}

func TestRedirectTrainlingSlashDisabled(t *testing.T) {
	m := New()
	m.RedirectTrailingSlash = false

	for i, tt := range []struct {
		method   string
		path     string
		code     int
		addToMux bool
	}{
		{"GET", "/foo", 200, true},
		{"GET", "/foo/", 404, false},
		{"GET", "/bar/", 200, true},
		{"POST", "/foobar", 200, true},
		{"POST", "/foobar/", 404, false},
	} {
		if tt.addToMux {
			m.add(tt.method, tt.path, h)
		}

		res := httptest.NewRecorder()
		m.ServeHTTP(res, newRequest(tt.method, tt.path, nil))
		if res.Code != tt.code {
			t.Errorf("[%d]: for path %q: got code %d; want %d", i, tt.path, res.Code, tt.code)
		}
	}
}

func TestMethodNotAllowed(t *testing.T) {
	m := New()
	m.Get("/bar", h)
	m.Post("/bar", h)

	res := httptest.NewRecorder()
	m.ServeHTTP(res, newRequest("PUT", "/bar", nil))
	if res.Code != http.StatusMethodNotAllowed {
		t.Errorf("for path %q: got code %d; want %d", "/bar", res.Code, http.StatusMethodNotAllowed)
	}

	got := res.Header().Get("Allow")
	want := strings.Join([]string{http.MethodHead, http.MethodGet, http.MethodPost}, ", ")
	if got != want {
		t.Errorf("got Allow header %v; want %v", got, want)
	}
}

func TestMethodNotAllowedDisabled(t *testing.T) {
	m := New()
	m.HandleMethodNotAllowed = false
	m.Get("/bar", h)
	m.Post("/bar", h)

	res := httptest.NewRecorder()
	m.ServeHTTP(res, newRequest("PUT", "/bar", nil))
	if res.Code != http.StatusNotFound {
		t.Errorf("for path %q: got code %d; want %d", "/bar", res.Code, http.StatusNotFound)
	}

	got := res.Header().Get("Allow")
	if got != "" {
		t.Errorf("got Allow header %v; want \"\"", got)
	}
}

func TestFormMethodFix(t *testing.T) {
	m := New()
	m.Get("/foo", h)
	m.Post("/foo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	m.Put("/foo", h)

	res := httptest.NewRecorder()
	req := newRequest(
		"POST",
		"/foo",
		strings.NewReader(url.Values{"_method": {"put"}}.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	m.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Errorf("for path %q: got code %d; want %d", "/foo", res.Code, http.StatusOK)
	}
}

func TestOtherMethods(t *testing.T) {
	m := New()
	m.Head("/bar", h)
	m.Delete("/bar", h)
	m.Options("/bar", h)
	m.Patch("/bar", h)

	res := httptest.NewRecorder()
	for _, method := range []string{"HEAD", "DELETE", "OPTIONS", "PATCH"} {
		m.ServeHTTP(res, newRequest(method, "/bar", nil))
		if res.Code != http.StatusOK {
			t.Errorf("for path %q: got code %d; want %d", "/bar", res.Code, http.StatusOK)
		}
	}

	m.ServeHTTP(res, newRequest("GET", "/bar", nil))
	if res.Code != http.StatusMethodNotAllowed {
		t.Errorf("for path %q: got code %d; want %d", "/bar", res.Code, http.StatusMethodNotAllowed)
	}
}

func TestNoBeginningSlash(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("the code did not panic")
		}
	}()

	m := New()
	m.Delete("bar", h)
}

func newRequest(method string, path string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		panic(err)
	}

	return req
}
