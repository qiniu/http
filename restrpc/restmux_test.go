package restrpc

import (
	"strings"
	"testing"
	"net/http"
)

type testcase struct {
	method  string
	path    string
	pattern string
	args    []string
	failed  bool
}

func matchAsNetHTTP(pattern string, urlPath string) bool {

	mux := http.NewServeMux()
	mux.Handle(pattern, mux)
	req, err := http.NewRequest("GET", "http://foo.com" + urlPath, nil)
	if err != nil {
		return false
	}
	_, match := mux.Handler(req)
	return match != ""
}

func patternMatch(pattern Pattern, method, urlPath string) (args []string, ok bool) {

	if pattern[0] == "" { // 用net/http的pattern方式
		ok = matchAsNetHTTP(pattern[1], urlPath)
	} else {
		parts := strings.Split(urlPath[1:], "/")
		args, ok = pattern.Match(method, parts)
	}
	return
}

var matchCases = []testcase{
	{
		method:  "POST",
		path:    "/bar",
		pattern: "POST /bar",
	},
	{
		method:  "GET",
		path:    "/bar/bar-param",
		pattern: "GET /bar/*",
		args:    []string{"bar-param"},
	},
	{
		method:  "DELETE",
		path:    "/bar/bar-param/foo",
		pattern: "DELETE /bar/*/foo",
		args:    []string{"bar-param"},
	},
	{
		method:  "POST",
		path:    "/bar/bar-param/foo/foo-param",
		pattern: "POST /bar/*/foo/*",
		args:    []string{"bar-param", "foo-param"},
	},
}

func TestMatch(t *testing.T) {

	for _, c := range matchCases {
		if args, ok := patternMatch(NewPattern(c.pattern), c.method, c.path); ok {
			if strings.Join(args, "/") != strings.Join(c.args, "/") {
				t.Fatal("pattern args not match =>", args, c.args)
			}
		} else if !c.failed {
			t.Fatal("not match =>", c)
		}
	}
}
