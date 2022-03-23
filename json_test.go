package snug_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
	"github.com/samharju/snug"
)

func TestMarshalError(t *testing.T) {
	its := is.New(t)
	c := make(chan int)
	res := httptest.NewRecorder()
	snug.JSON{"c": c}.Write(res, 200)
	its.Equal(res.Result().StatusCode, http.StatusInternalServerError)
}

func TestJSONWrite(t *testing.T) {
	its := is.New(t)
	res := httptest.NewRecorder()
	tc := snug.JSON{
		"a": 123,
		"b": "test",
		"c": []string{
			"a", "b", "c",
		},
	}
	tc.Write(res, http.StatusOK)

	its.Equal(res.Result().StatusCode, http.StatusOK)
	its.Equal(res.Header().Get("content-type"), "application/json")

	data, err := io.ReadAll(res.Result().Body)
	its.NoErr(err)
	its.Equal(
		string(data),
		`{"a":123,"b":"test","c":["a","b","c"]}`,
	)

}
