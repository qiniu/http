package hfac

import (
	"testing"
)

// ---------------------------------------------------------------------------

func TestPrefix(t *testing.T) {

	cases := [][2]string{
		{"Do", ""},
		{"Do123", "Do"},
		{"Dosomething", ""},
		{"DosomethingGood", "Dosomething"},
		{"Dosomething_Good", "Dosomething"},
	}

	for _, c := range cases {
		prefix, ok := prefixOf(c[0])
		if !ok {
			if c[1] != "" {
				t.Fatal("prefixOf bad case:", c[0], c[1])
			}
		} else {
			if c[1] != prefix {
				t.Fatal("prefixOf bad case:", c[0], c[1])
			}
		}
	}
}

func (r HandlerFactory) badUnion(r2 HandlerFactory) HandlerFactory {

	n := len(r2) - 1
	if n >= 0 && r2[n].Prefix == "Do" {
		r2 = r2[:n]
	}
	return append(r, r2...)
}

func TestUnion(t *testing.T) {

	f1 := HandlerFactory{
		{"Do", NewHandler},
	}
	f2 := HandlerFactory{
		{"Cmd", NewHandler},
		{"Do", NewHandler},
	}
	f3 := HandlerFactory{
		{"C1", NewHandler},
		{"C2", NewHandler},
		{"Do", NewHandler},
	}
	f4 := HandlerFactory{
		{"Cmd", NewHandler},
		{"C1", NewHandler},
		{"C2", NewHandler},
		{"Do", NewHandler},
	}

	r1 := f1.Union(f2)
	if f1[0].Prefix != "Do" {
		t.Fatal("Union change f1")
	}
	if len(r1) != 2 || !checkEqual(f2, r1) {
		t.Fatal("Union(f1, f2) != f2:", f2, r1)
	}

	r2 := f2.Union(f3)
	if !checkEqual(f4, r2) {
		t.Fatal("Union(f2, f3) != f4:", f4, r2)
	}
}

func checkEqual(a, b HandlerFactory) bool {

	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v.Prefix != b[i].Prefix {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
