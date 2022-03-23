package snug

import (
	"log"
	"net/http"
)

type logger struct {
	http.ResponseWriter
	status int
	size   int
}

// Wrap call to responsewriter to capture status code in logger.
func (l *logger) WriteHeader(s int) {
	l.ResponseWriter.WriteHeader(s)
	l.status = s
}

// Wrap call to responsewriter to capture response size in logger
func (l *logger) Write(b []byte) (int, error) {
	size, err := l.ResponseWriter.Write(b)
	l.size = size
	return size, err
}

// Log request info in a generic webserver way to stderr:
//
//	timestamp           | remote address             | method |status| response size | url path
//
// example:
//
//	2022/09/28 19:21:45 | 161.251.242.12:33706       | GET    | 200 |     56 | /items
//	2022/09/28 19:21:45 | 161.251.242.12:33706       | GET    | 405 |      0 | /favicon.ico
func Logging(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		w := &logger{ResponseWriter: rw}
		next(w, r)
		log.Printf("| %-21s | %-6s | %d | %6d | %s", r.RemoteAddr, r.Method, w.status, w.size, r.URL.String())
	}
}

// Middleware for catching panics.
// When handler panics, server logs the error and responds with status code 500 and body:
//
//	{"error": "internal server error"}
func Recover(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recover: panic: %s", r)
				JSON{"error": "internal server error"}.Write(w, http.StatusInternalServerError)
			}
		}()
		next(w, req)
	}
}
