package snug_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
	"github.com/samharju/snug"
)

// basic usecases covered
func TestRouter(t *testing.T) {
	its := is.New(t)

	testcases := []struct {
		name   string
		method string
		path   string

		reqMethod string
		reqPath   string

		expectedStatus    int
		expectedVariables int
	}{
		{
			"simple get",
			"GET", "/api/path",
			"GET", "/api/path", 200, 0,
		},
		{
			"simple put",
			"PUT", "/api/path",
			"PUT", "/api/path", 200, 0,
		},
		{
			"simple post",
			"POST", "/api/path",
			"POST", "/api/path", 200, 0,
		},
		{
			"simple delete",
			"DELETE", "/api/path",
			"DELETE", "/api/path", 200, 0,
		},
		{
			"variable in url",
			"GET", "api/<variable>",
			"GET", "/api/var", 200, 1,
		},
		{
			"two variables in url",
			"GET", "api/<variable>/<variable>",
			"GET", "/api/var/var", 200, 2,
		},
		{
			"variable midpath",
			"GET", "api/<variable>/_status",
			"GET", "/api/var/_status", 200, 1,
		},
		{
			"not found",
			"GET", "/api",
			"GET", "/api/_update", 404, 0,
		},
		{
			"not found if length matches",
			"GET", "/api/_update",
			"GET", "/api/_delete", 404, 0,
		},
		{
			"method not allowed",
			"GET", "/api",
			"POST", "/api", 405, 0,
		},
		{
			"wildcard pattern",
			"GET", "/api/*", "GET", "/api/anything/cool", 200, 0,
		},
		{
			"wildcard method",
			"*", "/api",
			"GET", "/api", 200, 0,
		},
		{
			"path len match 404",
			"GET", "/api",
			"POST", "/api2", 404, 0,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// params := []string{}
			f := func(rw http.ResponseWriter, r *http.Request) {
				// params = snug.UrlParams(uc.path, uc.reqPath)
				rw.WriteHeader(http.StatusOK)
			}
			r := snug.New()
			r.HandleFunc(tc.method, tc.path, f)
			rw := httptest.NewRecorder()
			req := httptest.NewRequest(tc.reqMethod, tc.reqPath, nil)

			r.ServeHTTP(rw, req)
			its.Equal(rw.Result().StatusCode, tc.expectedStatus) // actual status != expected
			// its.Equal(len(params), uc.expectedVariables)         // parsed url param count != expected count

		})
	}
}

func TestRegisterHandle(t *testing.T) {
	its := is.New(t)
	f := func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusAccepted)
	}
	r := snug.New()
	r.Handle("GET", "/", http.HandlerFunc(f))
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	r.ServeHTTP(rw, req)
	its.Equal(rw.Result().StatusCode, 202) // actual status != expected

}

func TestHttpRouterMethods(t *testing.T) {
	its := is.New(t)
	f := func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}
	r := snug.New()
	r.Get("/", f)
	r.Put("/", f)
	r.Post("/", f)
	r.Delete("/", f)
	rw := httptest.NewRecorder()
	get := httptest.NewRequest("GET", "/", nil)
	put := httptest.NewRequest("PUT", "/", nil)
	post := httptest.NewRequest("POST", "/", nil)
	delete := httptest.NewRequest("DELETE", "/", nil)

	reqs := []*http.Request{get, put, post, delete}
	for _, req := range reqs {

		r.ServeHTTP(rw, req)
		its.Equal(rw.Result().StatusCode, 200) // actual status != expected
	}
}

// router must be chainable to group handlers for shared middleware etc
func TestHandleSubrouter(t *testing.T) {
	its := is.New(t)

	f := func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusAccepted)
	}

	greetapi := snug.New()
	greetapi.Prefix = "/greet/"
	greetapi.HandleFunc("GET", "/apua", f)

	main := snug.New()
	main.Handle("*", "/greet/*", greetapi)

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/greet/apua", nil)

	main.ServeHTTP(rw, req)
	its.Equal(rw.Result().StatusCode, 202) // actual status != expected
}

func TestUrlparams(t *testing.T) {
	its := is.New(t)

	r := snug.New()
	r.UseMiddleware(snug.Logging)

	t.Run("parse url params", func(t *testing.T) {
		r.HandleFunc("GET", "/api/<good>/path/<best>/", func(w http.ResponseWriter, r *http.Request) {
			val := snug.Param(r, "good")
			its.Equal(val, "pizza")
			val = snug.Param(r, "best")
			its.Equal(val, "beer")
			val = snug.Param(r, "none")
			its.Equal(val, "")
			w.WriteHeader(200)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/pizza/path/beer", nil)

		r.ServeHTTP(rec, req)
		its.Equal(rec.Result().StatusCode, 200)
	})

	t.Run("no params does not crash", func(t *testing.T) {
		r.HandleFunc("GET", "/api/", func(w http.ResponseWriter, r *http.Request) {
			val := snug.Param(r, "any")
			its.Equal(val, "")
			w.WriteHeader(200)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/", nil)

		r.ServeHTTP(rec, req)
		its.Equal(rec.Result().StatusCode, 200)
	})

	t.Run("bad variable syntax handled", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				its.Equal(r, "malformed pattern: /api/<variable")
			}
		}()
		r.HandleFunc("GET", "/api/<variable", func(w http.ResponseWriter, r *http.Request) {})
	})

}

func TestDuplicateRoutePanics(t *testing.T) {

	its := is.New(t)

	defer func() {
		r := recover()
		its.Equal(r, "duplicate route and method: /test, GET")
	}()
	r := snug.New()
	r.HandleFunc("GET", "/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	r.HandleFunc("GET", "/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

}

// just a benchmark to check against http.ServeMux
func BenchmarkSnug(b *testing.B) {
	r := snug.New()
	r.HandleFunc("GET", "/test/<variable>", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("test"))
	})

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/hello/asd", nil)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r.ServeHTTP(rw, req)
	}
}

func BenchmarkHttplib(b *testing.B) {
	r := http.NewServeMux()
	r.HandleFunc("/test", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("test"))
	})

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/hello", nil)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r.ServeHTTP(rw, req)
	}
}
