package strconv

import (
	"testing"
)

type testCase struct {
	v         interface{}
	omitempty bool
	ret       string
	err       error
}

func TestEncode(t *testing.T) {

	cases := []testCase{
		{0, true, "", ErrOmit},
		{0, false, "0", nil},
		{1.2, true, "1.2", nil},
		{1.2, false, "1.2", nil},
		{false, false, "false", nil},
		{false, true, "", ErrOmit},
		{"abc", true, "abc", nil},
		{"", false, "", nil},
		{"", true, "", ErrOmit},
	}

	for _, c := range cases {
		ret, err := Encode(c.v, c.omitempty)
		if ret != c.ret || err != c.err {
			t.Fatal("Encode failed:", c)
		}
	}
}
