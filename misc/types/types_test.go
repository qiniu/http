package types

import (
	"reflect"
	"testing"
)

// --------------------------------------------------------------------

func isType(v interface{}, cate Category) bool {

	return Is(reflect.ValueOf(v).Kind(), cate)
}

func Test(t *testing.T) {

	var a int
	var b int8
	var c float32

	if !isType(a, Ints) || !isType(a, Int) || isType(a, Int8) {
		t.Fatal("Test int failed")
	}

	if !isType(b, Ints) || isType(b, Int) || !isType(b, Int8) {
		t.Fatal("Test int8 failed")
	}

	if isType(c, Ints) || !isType(c, Floats) || !isType(c, Float32) {
		t.Fatal("Test float32 failed")
	}
}

// --------------------------------------------------------------------
