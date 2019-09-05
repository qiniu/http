package types

import (
	"reflect"
)

// --------------------------------------------------------------------

type Category uint

const (
	Invalid Category = 1 << iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Ptr
	Slice
	String
	Struct
	UnsafePointer

	Ints     = Int | Int8 | Int16 | Int32 | Int64
	Uints    = Uint | Uint8 | Uint16 | Uint32 | Uint64 | Uintptr
	Floats   = Float32 | Float64
	Complexs = Complex64 | Complex128
)

func Is(kind reflect.Kind, cate Category) bool {

	return ((1 << kind) & cate) != 0
}

// --------------------------------------------------------------------
