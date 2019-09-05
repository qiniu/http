package restrpc

import (
	"fmt"
	"strings"
	"testing"
)

type patternTestCase struct {
	Method  string
	Pattern Pattern
	Sep     string
}

func TestPatternOf(t *testing.T) {

	cases := []patternTestCase{
		{"AppleBanana", Pattern{"Apple", "Banana"}, "_"},
		{"Apple_Banana", Pattern{"Apple", "*", "Banana"}, "_"},
		{"AppleBanana_", Pattern{"Apple", "Banana", "*"}, "_"},
		{"Apple_Banana_", Pattern{"Apple", "*", "Banana", "*"}, "_"},
		{"AppleBanana", Pattern{"Apple", "Banana"}, "__"},
		{"App_le__Banana", Pattern{"App_le", "*", "Banana"}, "__"},
		{"Apple_Banana__", Pattern{"Apple_", "Banana", "*"}, "__"},
		{"Apple__Banana__", Pattern{"Apple", "*", "Banana", "*"}, "__"},
	}
	for _, c := range cases {
		pattern := patternOf(c.Method, c.Sep)
		fmt.Println("patternOf:", c.Method, pattern)
		if strings.Join(pattern, "/") != strings.Join(c.Pattern, "/") {
			t.Fatal("patternOf failed:", c, pattern)
		}
	}
}
