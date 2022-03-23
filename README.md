# Snug Minimal HTTP JSON Api Framework

*snug: fitting closely and comfortably: "a snug coat"* - [Merriam-Webster](https://www.merriam-webster.com/dictionary/snug)

Snug is a simple router with the goal of having the bare minimum to put together a JSON web API.

Writing small web apis with Go is barely a chore, it's a pleasure.
Repeating the same few simple steps over and over again, I decided to put together this package
mostly for quick prototyping. And just to build a router from scratch and do something with
`reflect`.

Provided features:
- Router with minimal functionalities mimicking `http.ServeMux`
- Some default error responses
- Request body binding with `snug.Fit`
- Logging
- `snug.JSON` for dumping simple json responses to responsewriter


Example with all the bells and whistles in place:
```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/samharju/snug"
)

type server struct {
	router *snug.Router
	db     map[string]int
}

func main() {
	mux := snug.New()

	// init a server with router and a db
	srv := server{
		router: mux,
		db: map[string]int{
			"esa": 25,
		},
	}

	// add logging
	mux.UseMiddleware(snug.Recover)
	mux.UseMiddleware(snug.Logging)

	// register handlers
	mux.Get("/greet", srv.greetanon)
	mux.Post("/greet", srv.postgreet)
	mux.Get("/greet/age/<name>", srv.getAge)

	log.Fatalln(http.ListenAndServe(":8000", mux))
}

// handle url parameter
func (s server) getAge(w http.ResponseWriter, r *http.Request) {
	name := snug.Param(r, "name")
	age, ok := s.db[name]

	if !ok {
		snug.JSON{"error": name + " not found"}.Write(w, 404)
		return
	}

	snug.JSON{"age": age}.Write(w, 200)

}

// endpoint with no input
func (server) greetanon(w http.ResponseWriter, r *http.Request) {
	snug.JSON{"msg": "hello anon"}.Write(w, 200)
}

// reading post request body
func (server) postgreet(w http.ResponseWriter, r *http.Request) {

	var req struct {
		Name *string `json:"name" snug:"required"`
		Age  *int    `json:"age" snug:"required"`
	}

	err := snug.Fit(r.Body, &req)
	if err != nil {
		// return 400 if missing required fields
		snug.JSON{"error": err.Error()}.Write(w, 400)
		return
	}

	snug.JSON{
		"msg": fmt.Sprintf("hello %s %d years", *req.Name, *req.Age),
	}.Write(w, 200)
}

```
