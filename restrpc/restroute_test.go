package restrpc_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/qiniu/http/restrpc"
)

// ---------------------------------------------------------------------------

type Service struct{}

type FooArgs struct {
	FormArg1 string `json:"a"`
	FormArg2 string `json:"b"`
}

func (r *Service) PostFoo(args *FooArgs, env *restrpc.Env) {
	io.WriteString(env.W, "PostFoo: "+env.Req.URL.String()+" Method: "+env.Req.Method)
}

type Foo_Args struct {
	FormArg1 string `json:"a"`
	FormArg2 string `json:"b"`
}

func (r *Service) PostFoo_(args *Foo_Args, env *restrpc.Env) {
	io.WriteString(env.W, "PostFoo_: "+env.Req.URL.String()+" Method: "+env.Req.Method)
}

type Foo_BarArgs struct {
	FormArg1 string `json:"a"`
	FormArg2 string `json:"b"`
}

func (r *Service) PostFoo_Bar(args *Foo_BarArgs, env *restrpc.Env) {
	io.WriteString(env.W, "PostFoo_Bar: "+env.Req.URL.String()+" Method: "+env.Req.Method)
}

func (r *Service) PutFoo_Bar(args *Foo_BarArgs, env *restrpc.Env) {
	io.WriteString(env.W, "PutFoo_Bar: "+env.Req.URL.String()+" Method: "+env.Req.Method)
}

func (r *Service) DeleteFoo_Bar(args *Foo_BarArgs, env *restrpc.Env) {
	io.WriteString(env.W, "DeleteFoo_Bar: "+env.Req.URL.String()+" Method: "+env.Req.Method)
}

func (r *Service) GetFoo_Bar(args *Foo_BarArgs, env *restrpc.Env) {
	io.WriteString(env.W, "GetFoo_Bar: "+env.Req.URL.String()+" Method: "+env.Req.Method)
}

func (r *Service) Default(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Do: "+req.URL.Path)
}

type Foo_Bar_Args struct {
	FormArg1 string `json:"a"`
	FormArg2 string `json:"b"`
}

func (r *Service) PostFoo_Bar_(args *Foo_Bar_Args, env *restrpc.Env) {
	io.WriteString(env.W, "PostFoo_Bar_: "+env.Req.URL.String()+" Method: "+env.Req.Method)
}

type Apple_Banana_Args struct {
	FormArg1 string `json:"a"`
	FormArg2 string `json:"b"`
}

func (r *Service) PostApple_Banana_(args *Apple_Banana_Args, env *restrpc.Env) (ret Apple_Banana_Args, err error) {
	return *args, nil
}

type Banana_Apple_Args struct {
	ReqBody []struct {
		A int `json:"a"`
	}
}

func (r *Service) PostBanana_Apple_(args *Banana_Apple_Args, env *restrpc.Env) (ret Banana_Apple_Args, err error) {
	return *args, nil
}

// ---------------------------------------------------------------------------

var routeCases = [][3]string{
	{"POST", "http://localhost:2358/v1/foo?a=1&b=2", "PostFoo: /v1/foo?a=1&b=2 Method: POST"},
	{"POST", "http://localhost:2358/v1/foo/cmd1?a=1&b=2", "PostFoo_: /v1/foo/cmd1?a=1&b=2 Method: POST"},

	{"POST", "http://localhost:2358/v1/foo/cmd1/bar?a=1&b=2", "PostFoo_Bar: /v1/foo/cmd1/bar?a=1&b=2 Method: POST"},
	{"GET", "http://localhost:2358/v1/foo/cmd1/bar?a=1&b=2", "GetFoo_Bar: /v1/foo/cmd1/bar?a=1&b=2 Method: GET"},
	{"DELETE", "http://localhost:2358/v1/foo/cmd1/bar?a=1&b=2", "DeleteFoo_Bar: /v1/foo/cmd1/bar?a=1&b=2 Method: DELETE"},
	{"PUT", "http://localhost:2358/v1/foo/cmd1/bar?a=1&b=2", "PutFoo_Bar: /v1/foo/cmd1/bar?a=1&b=2 Method: PUT"},

	{"POST", "http://localhost:2358/v1/foo/cmd1/bar/cmd2?a=1&b=2", "PostFoo_Bar_: /v1/foo/cmd1/bar/cmd2?a=1&b=2 Method: POST"},

	{"POST", "http://localhost:2358/v1/do/cmd1/bar/cmd2?a=1&b=2", "Do: /v1/do/cmd1/bar/cmd2"},

	{"POST", "http://localhost:2358/v1/apple/cmd1/banana/cmd2?a=1&b=2", `{"a":"1","b":"2"}`},
}

func TestRoute(t *testing.T) {

	go func() {
		service := new(Service)
		router := restrpc.Router{
			PatternPrefix: "/v1",
			Default:       http.HandlerFunc(service.Default),
		}
		t.Fatal(router.ListenAndServe(":2358", service))
	}()
	time.Sleep(.5e9)

	var err error
	var cookies []*http.Cookie
	var resp *http.Response
	for _, c := range routeCases {
		req, _ := http.NewRequest(c[0], c[1], nil)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		resp, err = http.DefaultClient.Do(req)
		cookies = checkResp(t, resp, err, c[2])
	}
}

func TestJsonRouteWithOnlyCmdArgs(t *testing.T) {

	go func() {
		service := new(Service)
		router := restrpc.Router{
			PatternPrefix: "/v1",
			Default:       http.HandlerFunc(service.Default),
		}
		t.Fatal(router.ListenAndServe(":2359", service))
	}()
	time.Sleep(.5e9)

	var err error
	var cookies []*http.Cookie
	var resp *http.Response

	b := bytes.NewBuffer([]byte(`[{"a":1}]`))

	req, _ := http.NewRequest("POST", "http://localhost:2359/v1/banana/cmd1/apple/cmd2", b)
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	resp, err = http.DefaultClient.Do(req)
	checkResp(t, resp, err, `{"ReqBody":[{"a":1}]}`)
}

func checkResp(t *testing.T, resp *http.Response, err error, respText string) (cookies []*http.Cookie) {

	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if details, ok := resp.Header["X-Log"]; ok {
		for i, detail := range details {
			log.Println("Detail:", i, detail)
		}
	}

	text1, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("ioutil.ReadAll failed:", err)
	}

	text := string(text1)
	if text != respText {
		t.Fatal("unexpected resp:", text, "respExpected:", respText)
	}

	cookies = resp.Cookies()
	if len(cookies) != 0 {
		log.Println("Cookies:", cookies)
	}
	return
}

// ---------------------------------------------------------------------------
