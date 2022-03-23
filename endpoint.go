package snug

import (
	"net/http"
	"strings"
)

type endpoint struct {
	method   string
	pattern  string
	path     []string
	handler  http.HandlerFunc
	wildcard bool
	params   bool
}

func (e endpoint) match(method string, path []string) (hf http.HandlerFunc, notallowed, params bool) {
	var hit bool = true
	// url part count matches
	if (len(path) != len(e.path)) && !e.wildcard {
		return
	}

	// match parts
	for i, p := range e.path {
		// wildcard at any point matches
		if p == "*" {
			break
		}
		// check url part matches path part, variables pass
		if path[i] != p && !strings.HasPrefix(p, "<") {
			hit = false
			break
		}
	}
	// path pattern did not match
	if !hit {
		return
	}
	// method does not match endpoint
	if (e.method != "*") && (method != e.method) {
		notallowed = true
		return
	}
	// return handlerfunc and flag if endpoint has url parameters
	return e.handler, false, e.params
}
