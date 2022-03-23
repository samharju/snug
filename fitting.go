package snug

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Decode response body to given struct pointer.
//
// Error returned is either for invalid json or missing a value for field marked as required.
// Using snug-tag requires to use also json-tag.
//
// Due to restrictions in encoding/json, use a pointer type for fields that can hold
// a falsy value. For example empty string or number of value 0 is not interpreted as missing.
// Using pointers will provide a nil value to return an error if field is tagged as required.
//
//	func postgreet(w http.ResponseWriter, r *http.Request) {
//		var body struct {
//			Name *string `json:"name" snug:"required"`
//			Age  *int    `json:"age" snug:"required"`
//		}
//		err := snug.Fit(r.Body, &body)
//		if err != nil {
//			// use error for details
//			return
//		}
//		w.Write(200)
//	}
func Fit(data io.ReadCloser, o any) error {
	err := json.NewDecoder(data).Decode(o)
	if err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	t := reflect.TypeOf(o).Elem()
	v := reflect.ValueOf(o).Elem()
	errMsgs := []string{}
	for i := 0; i < t.NumField(); i++ {
		tags, ok := t.Field(i).Tag.Lookup("snug")
		if !ok {
			continue
		}
		jtags, ok := t.Field(i).Tag.Lookup("json")
		if !ok {
			panic("missing tag: json")
		}
		fieldName := strings.Split(jtags, ",")[0]
		tagList := strings.Split(tags, ",")
		for _, tag := range tagList {
			if tag == "required" && v.Field(i).IsZero() {
				errMsgs = append(errMsgs, fmt.Sprintf("'%s' is a required field", fieldName))
			}

		}
	}
	if len(errMsgs) != 0 {
		return fmt.Errorf(strings.Join(errMsgs, ", "))
	}
	return nil
}
