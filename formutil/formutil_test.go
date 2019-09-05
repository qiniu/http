package formutil

import (
	"fmt"
	"net/url"
	"testing"
)

// --------------------------------------------------------------------

type Float float32

type Foo struct {
	A    int      `json:"a"`
	B    []Float  `json:"b"`
	C    []string `json:"c,omitempty,has"`
	D    int      `json:"d,has"`
	E    int      `json:"e,default"`
	F    *int     `json:"f"`
	G    *int     `json:"g"`
	HasC bool
	HasD bool
}

func Test(t *testing.T) {

	form := url.Values{
		"a": {"-1"},
		"b": {"1.2", "3.4"},
		"c": {"abc", "df"},
		"f": {"-2"},
	}

	var ret Foo
	ret.D = 100
	ret.E = 100
	err := Parse(&ret, form)
	if err != nil {
		t.Fatal(ret, err)
	}
	fmt.Println(ret)
	if ret.D != 0 {
		t.Fatal("ret.D is not 0")
	}
	if ret.E != 100 {
		t.Fatal("ret.E is not 100")
	}
	if !ret.HasC || ret.HasD {
		t.Fatal("!ret.HasC || ret.HasD")
	}
	if ret.F == nil || *ret.F != -2 {
		t.Fatal("ret.F is not -2")
	}
	if ret.G != nil {
		t.Fatal("ret.G is not nil")
	}
}

// --------------------------------------------------------------------
