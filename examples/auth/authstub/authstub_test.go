package authstub

import (
	"fmt"
	"testing"
)

// --------------------------------------------------------------------

func TestFormatAndParse(t *testing.T) {

	auth := `QiniuStub uid=1&ut=2&app=3&suid=5&sut=6&ak=777%40&eu=%40x`
	user, err := Parse(auth)
	if err != nil {
		t.Fatal("Parse failed:", err)
	}

	auth2 := Format(&user)
	fmt.Println("auth:", auth2)
	fmt.Printf("user: %#v\n", user)

	if auth != auth2 {
		t.Fatal("Parse failed: auth != auth2")
	}
}

// --------------------------------------------------------------------
