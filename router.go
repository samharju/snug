package snug

import (
	"context"
	"log"
	"net/http"
	"strings"
)

// Initialize new snug router.
func New() *Router {
	return &Router{
		routes:           []endpoint{},
		NotFound:         func(w http.ResponseWriter, r *http.Request) { JSON{"error": "not found"}.Write(w, 404) },
		MethodNotAllowed: func(w http.ResponseWriter, r *http.Request) { JSON{"error": "method not allowed"}.Write(w, 405) },
	}
}

type Router struct {
	Prefix     string
	routes     []endpoint
	middleware []Middleware
	// NotFound is called when no route matches url.
	// Default handler returns status 404 and a response body:
	// {"error": "not found"}
	NotFound http.HandlerFunc
	// MethodNotAllowed is called when a pattern matches but method for that pattern does not match.
	// Default handler returns status 405 with no body.
	MethodNotAllowed http.HandlerFunc
}

// Middleware accepts a http.HandlerFunc and returns a http.HandlerFunc.
// Perform middleware things before and after calling wrapped handler.
//
//	func SomeMiddleware(next http.HandlerFunc) http.HandlerFunc {
//		// do things here when handler is registered to a route
//		return func(rw http.ResponseWriter, r *http.Request) {
//			// do things here before calling next
//			next(w, r)
//			// do things here after next returns
//		}
//	}
type Middleware func(http.HandlerFunc) http.HandlerFunc

// Slap given middleware to all handlers to be registered on this router.
//
//		r := snug.New()
//		mw := func(hf http.HandlerFunc) http.HandlerFunc {
//			return func(rw http.ResponseWriter, r *http.Request) {
//				fmt.Println("middleware called")
//				hf(rw, r)
//			}
//		}
//		r.UseMiddleware(mw)
//
//		r.HandleFunc("GET", "/path", func(rw http.ResponseWriter, r *http.Request) {
//			fmt.Println("handler called")
//		})
//
//	 // GET /path
//	 // middleware called
//	 // handler called
func (r *Router) UseMiddleware(mw Middleware) {
	r.NotFound = mw(r.NotFound)
	r.MethodNotAllowed = mw(r.MethodNotAllowed)
	r.middleware = append(r.middleware, mw)
}

func urlParams(pattern, path string) (params map[string]string) {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	patternParts := strings.Split(strings.ToLower(strings.Trim(pattern, "/")), "/")
	params = map[string]string{}
	for i, p := range patternParts {
		if strings.HasPrefix(p, "<") {
			params[strings.Trim(p, "<>")] = pathParts[i]
		}
	}
	return
}

type contextVar string

// Pull url parameter with name.
//
//	r.HandleFunc("GET", "/api/<good>/path/<best>/", func(w http.ResponseWriter, r *http.Request) {
//		// Request path /api/pizza/path/beer
//		val1 := snug.Param(r, "good") // pizza
//		val2 := snug.Param(r, "best") // beer
//	})
func Param(r *http.Request, key string) string {
	value, ok := r.Context().Value(contextVar("params")).(map[string]string)
	if !ok {
		return ""
	}
	return value[key]
}

// ServeHTTP routes request to appropriate handler.
func (ro Router) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var (
		hf         http.HandlerFunc
		notallowed bool
		hasParams  bool
	)
	path := strings.Split(strings.ToLower(strings.Trim(r.URL.Path, "/")), "/")

	for _, v := range ro.routes {
		hf, notallowed, hasParams = v.match(r.Method, path)
		if hf != nil {
			if hasParams {
				params := urlParams(v.pattern, r.URL.Path)
				r = r.WithContext(context.WithValue(r.Context(), contextVar("params"), params))
			}

			hf.ServeHTTP(rw, r)
			return
		}
	}
	if notallowed {
		ro.MethodNotAllowed.ServeHTTP(rw, r)
		return
	}
	ro.NotFound.ServeHTTP(rw, r)
}

func parsePattern(pattern string) (bool, []string) {
	p := strings.Split(strings.ToLower(strings.Trim(pattern, "/")), "/")

	for i := range p {
		pre := strings.HasPrefix(p[i], "<")
		suf := strings.HasSuffix(p[i], ">")
		if pre || suf {
			if !(pre && suf) {
				panic("malformed pattern: " + pattern)
			}
			return true, p
		}
	}
	return false, p
}

// Register http.HandleFunc to given method and path.
func (r *Router) HandleFunc(method, path string, f http.HandlerFunc) {
	if r.Prefix != "" {
		path = strings.Trim(r.Prefix, "/") + "/" + strings.TrimLeft(path, "/")
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	hasparams, p := parsePattern(path)

	for _, v := range r.routes {
		hf, _, _ := v.match(method, p)
		if hf != nil {
			panic("duplicate route and method: " + path + ", " + method)
		}
	}

	for _, mw := range r.middleware {
		f = mw(f)
	}
	e := endpoint{
		method:  strings.ToUpper(method),
		pattern: path,
		path:    p,
		handler: f,
		params:  hasparams,
	}
	if p[len(p)-1] == "*" {
		e.wildcard = true
	}
	r.routes = append(r.routes, e)
	log.Printf("Serving %s %s", method, path)
}

// Register http.Handler to given method and path.
func (r *Router) Handle(method, path string, h http.Handler) {
	r.HandleFunc(method, path, h.ServeHTTP)
}

// Register handler to path with GET.
func (r *Router) Get(path string, hf http.HandlerFunc) {
	r.HandleFunc("GET", path, hf)
}

// Register handler to path with POST.
func (r *Router) Post(path string, hf http.HandlerFunc) {
	r.HandleFunc("POST", path, hf)
}

// Register handler to path with PUT.
func (r *Router) Put(path string, hf http.HandlerFunc) {
	r.HandleFunc("PUT", path, hf)
}

// Register handler to path with DELETE.
func (r *Router) Delete(path string, hf http.HandlerFunc) {
	r.HandleFunc("DELETE", path, hf)
}
