package snug

import (
	"encoding/json"
	"log"
	"net/http"
)

// Add Content-Type: application/json-header to response.
func AddJsonHeader(rw http.ResponseWriter) {
	rw.Header().Add("Content-Type", "application/json")
}

// JSON type enables easy dumping of contents to http.ResponseWriter.
type JSON map[string]any

func (s JSON) dump() ([]byte, error) {
	serialized, err := json.Marshal(s)
	if err != nil {
		log.Printf("ERROR: JSON.Dump: %s", err)
		return []byte(`{"error": "internal server error"}`), err
	}
	return serialized, nil
}

// Write map as a json payload to responsewriter.
//
// Adds content-type header and given status code.
// If serialization fails, returns a status 500 and a message:
//
//	{"error": "internal server error"}
//
// Dump some fields to JSON and write to http.ResponseWriter:
//
//	func hello(w http.ResponseWriter, r *http.Request) {
//		m := snug.JSON{
//			"msg":   "hello",
//			"some":  "field",
//			"other": 123,
//		}
//		m.Write(w, 200)
//	}
func (s JSON) Write(w http.ResponseWriter, status int) {
	AddJsonHeader(w)
	res, err := s.dump()
	if err != nil {
		w.WriteHeader(500)
	} else {
		w.WriteHeader(status)
	}
	w.Write(res)
}
