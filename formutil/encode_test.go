package formutil

import (
	"testing"
)

// --------------------------------------------------------------------

type encodeTestCase struct {
	v   interface{}
	ret string
	err error
}

type foo struct {
	A int      `json:"a,omitempty"`
	B []int    `json:"b,omitempty"`
	C []string `json:"c,omitempty"`
	D bool     `json:"d"`
}

func TestEncode(t *testing.T) {

	cases := []encodeTestCase{
		{&foo{A: 1}, "a=1&d=false", nil},
		{&foo{A: 0, D: true}, "d=true", nil},
		{&foo{A: 0, B: []int{1, 3, 2}, D: true}, "b=1&b=3&b=2&d=true", nil},
		{&foo{A: 0, C: []string{"1 3", "a4", "b2"}, D: true}, "c=1+3&c=a4&c=b2&d=true", nil},
	}

	for _, c := range cases {
		ret, err := EncodeToString(c.v)
		if ret != c.ret || err != c.err {
			t.Fatal("EncodeToString failed:", c, ret, err)
		}
	}
}

// --------------------------------------------------------------------
