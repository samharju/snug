package snug_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"
	"github.com/samharju/snug"
)

// middleware does what middleware does
func TestMiddleware(t *testing.T) {
	its := is.New(t)

	callcount := 0

	mw := func(next http.HandlerFunc) http.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request) {
			callcount += 1
			next(rw, r)
		}
	}

	r := snug.New()
	r.UseMiddleware(mw)
	r.HandleFunc("GET", "/test1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	r.HandleFunc("GET", "/test2", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test1", nil)
	req2 := httptest.NewRequest("GET", "/test2", nil)

	r.ServeHTTP(rec, req)
	r.ServeHTTP(rec, req2)

	its.Equal(callcount, 2) // middleware was not called on both handlers

}

func TestLogging(t *testing.T) {
	its := is.New(t)

	var buf bytes.Buffer
	log.SetOutput(&buf)

	r := snug.New()
	r.UseMiddleware(snug.Logging)
	r.HandleFunc("GET", "/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("test"))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	r.ServeHTTP(rec, req)
	its.Equal(rec.Result().StatusCode, 200)
	its.True(strings.HasSuffix(buf.String(), "| GET    | 200 |      4 | /\n")) // not sure what was logged

}

func TestRecover(t *testing.T) {
	its := is.New(t)

	var buf bytes.Buffer
	log.SetOutput(&buf)

	r := snug.New()
	r.UseMiddleware(snug.Recover)
	r.HandleFunc("GET", "/", func(w http.ResponseWriter, r *http.Request) {
		panic("test")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	r.ServeHTTP(rec, req)
	its.Equal(rec.Result().StatusCode, http.StatusInternalServerError)

	response, err := io.ReadAll(rec.Result().Body)
	its.NoErr(err)
	its.Equal(string(response), `{"error":"internal server error"}`)
	fmt.Println(buf.String(), "test")

}
