package snug_test

import (
	"io"
	"strings"
	"testing"

	"github.com/matryer/is"
	"github.com/samharju/snug"
)

func strreadcloser(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}

func TestFitRequired(t *testing.T) {

	type teststruct struct {
		A int               `json:"a" snug:"required"`
		B string            `json:"b" snug:"required"`
		C []string          `json:"c" snug:"required"`
		D map[string]string `json:"d" snug:"required"`
		E *int              `json:"e" snug:"required"`
		F bool              `json:"f" snug:"required"`
		G *bool             `json:"g" snug:"required"`
		H string
	}

	type testcase struct {
		name      string
		input     string
		errString string
	}

	testcases := []testcase{
		{
			"invalid json",
			`{"adf`,
			"invalid json: unexpected EOF",
		},
		{
			"required int",
			`{"b": "string", "c": ["string"], "d": {"string": "string"}, "e": 1, "f": true, "g": false}`,
			"'a' is a required field",
		},
		{
			"required string",
			`{"a":1, "c": ["string"], "d": {"string": "string"}, "e": 1, "f": true, "g": false}`,
			"'b' is a required field",
		},
		{
			"required string array",
			`{"a":1, "b": "string", "d": {"string": "string"}, "e": 1, "f": true, "g": false}`,
			"'c' is a required field",
		},
		{
			"required object",
			`{"a":1, "b": "string", "c": ["string"], "e": 1, "f": true, "g": false}`,
			"'d' is a required field",
		},
		{
			"required int with 0 as valid value",
			`{"a":1, "b": "string", "c": ["string"], "d": {"string": "string"}, "f": true, "g": false}`,
			"'e' is a required field",
		},
		{
			"required bool accepting only true",
			`{"a":1, "b": "string", "c": ["string"], "d": {"string": "string"}, "e": 1, "g": false}`,
			"'f' is a required field",
		},
		{
			"required bool any value",
			`{"a":1, "b": "string", "c": ["string"], "d": {"string": "string"}, "e": 1, "f": true}`,
			"'g' is a required field",
		},
	}

	its := is.New(t)

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var o teststruct
			err := snug.Fit(strreadcloser(tc.input), &o)
			its.True(err != nil)                 // expected err but got nil
			its.Equal(err.Error(), tc.errString) // got != excpected
		})
	}

	t.Run("no errors", func(t *testing.T) {
		type tc struct {
			A string `json:"a" snug:"required"`
			B int    `json:"b" snug:"required"`
		}
		tci := tc{}
		snug.Fit(strreadcloser(`{"a": "abc", "b": 123}`), &tci)
		its.Equal(tci.A, "abc")
		its.Equal(tci.B, 123)

	})

}

func TestMissingJsonTag(t *testing.T) {
	its := is.New(t)

	type testcase struct {
		a string `snug:"required"`
	}

	tc := testcase{a: "123"}
	defer func() {
		if r := recover(); r != nil {
			its.Equal(r, "missing tag: json")
		}
	}()
	snug.Fit(strreadcloser(`{"a": "abc"}`), &tc)

}
