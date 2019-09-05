package strconv

import (
	"bytes"
	"testing"
	"time"
)

// --------------------------------------------------------------------

type MyTime struct {
	time.Time
}

func (p *MyTime) ParseValue(str string) (err error) {
	p.Time, err = time.Parse("20060102", str)
	return
}

func TestCustom(t *testing.T) {

	var tv MyTime
	err := Parse(&tv, "20130920")
	if err != nil {
		t.Fatal(tv, err)
	}
	y, m, d := tv.Date()
	if y != 2013 || m != 9 || d != 20 {
		t.Fatal(y, m, d)
	}
}

// --------------------------------------------------------------------

func TestBool(t *testing.T) {

	var bv bool
	err := Parse(&bv, "true")
	if err != nil || bv != true {
		t.Fatal(bv, err)
	}

	err = Parse(&bv, "True")
	if err != nil || bv != true {
		t.Fatal(bv, err)
	}

	err = Parse(&bv, "1")
	if err != nil || bv != true {
		t.Fatal(bv, err)
	}

	err = Parse(&bv, "2")
	if err == nil {
		t.Fatal(bv, err)
	}

	err = Parse(&bv, "TRUE")
	if err != nil || bv != true {
		t.Fatal(bv, err)
	}

	var pbv *bool
	err = Parse(&pbv, "TRUE")
	if err != nil || pbv == nil || *pbv != true {
		t.Fatal(pbv, err)
	}
}

// --------------------------------------------------------------------

type Integer int64

func TestInt(t *testing.T) {

	{
		var v int
		err := Parse(&v, "1.2")
		if err == nil {
			t.Fatal(v, err)
		}

		err = Parse(&v, "-1")
		if err != nil || v != -1 {
			t.Fatal(v, err)
		}
	}
	{
		var v Integer
		err := Parse(&v, "1.2")
		if err == nil {
			t.Fatal(v, err)
		}

		err = Parse(&v, "-1")
		if err != nil || v != -1 {
			t.Fatal(v, err)
		}
	}
}

// --------------------------------------------------------------------

type Uinteger uint32

func TestUint(t *testing.T) {

	{
		var v uint
		err := Parse(&v, "1.2")
		if err == nil {
			t.Fatal(v, err)
		}

		err = Parse(&v, "1")
		if err != nil || v != 1 {
			t.Fatal(v, err)
		}

		err = Parse(&v, "-1")
		if err == nil {
			t.Fatal(v, err)
		}
	}
	{
		var v Uinteger
		err := Parse(&v, "1.2")
		if err == nil {
			t.Fatal(v, err)
		}

		err = Parse(&v, "1")
		if err != nil || v != 1 {
			t.Fatal(v, err)
		}

		err = Parse(&v, "-1")
		if err == nil {
			t.Fatal(v, err)
		}
	}
}

// --------------------------------------------------------------------

type ByteSlice []byte

func TestBytes(t *testing.T) {

	{
		var v []byte
		err := Parse(&v, "1.2")
		if err != nil {
			t.Fatal(v, err)
		}

		if !bytes.Equal(v, []byte{'1', '.', '2'}) {
			t.Fatal(v)
		}
	}
	{
		var v ByteSlice
		err := Parse(&v, "1.2")
		if err != nil {
			t.Fatal(v, err)
		}

		if !bytes.Equal(v, []byte{'1', '.', '2'}) {
			t.Fatal(v)
		}
	}
}

// --------------------------------------------------------------------
