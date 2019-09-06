package authrestrpc

import (
	"testing"

	"github.com/qiniu/http/restrpc"
	"github.com/qiniu/qiniutest/httptest"
	"github.com/qiniu/x/log"
	"github.com/qiniu/x/mockhttp"
)

func init() {
	log.SetOutputLevel(1)
}

// ---------------------------------------------------------------------------

func TestServer(t *testing.T) {

	cfg := &Config{}

	svr, err := New(cfg)
	if err != nil {
		t.Fatal("New service failed:", err)
	}

	transport := mockhttp.NewTransport()
	router := restrpc.Router{
		PatternPrefix: "/v1",
		Mux:           restrpc.NewServeMux(),
	}
	transport.ListenAndServe("192.168.0.10:9898", router.Register(svr))

	ctx := httptest.New(t)
	ctx.SetTransport(transport)

	ctx.Exec(
		`
	host foo.com 192.168.0.10:9898
	auth testuser1 |authstub -uid 1 -utype 4|
	auth testuser2 |authstub -uid 2 -utype 4|

	#this is a comment
	#
	post http://foo.com/v1/foo/foo123/bar
	json '{"a": "1", "b": "2"}'
	ret 401
	json '{"error":"bad token"}'

	post http://foo.com/v1/foo/foo123/bar
	auth testuser1
	json '{"a": "1", "b": "2"}'
	ret 200
	json '{"id": $(id)}'

	get http://foo.com/v1/foo/$(id)
	auth testuser2
	ret 404
	json '{
		"error": "id not found"
	}'

	get http://foo.com/v1/foo/$(id)
	auth testuser1
	ret 200
	json '{"id": $(id), "foo": "foo123", "a": "1", "b": "2"}'

	get http://foo.com/v1/foo/1.3
	auth testuser1
	ret 404
	json '{
		"error": "id not found"
	}'

	match $(abcd) 4578
	println \n|base64 $(abcd)|
	post http://foo.com/v1/foo/|base64 $(abcd)|/bar
	auth testuser1
	form a=$(id)&b=3
	ret 200
	json '{
		"id": $(id2)
	}'

	post http://foo.com/v1/hosts/192.168.3.1
	auth testuser1
	json '{
		"foo.com": "127.0.0.1",
		"bar.com": "192.168.4.10"
	}'
	ret 200
	echo $(resp.body)
	json '{
		"CmdArgs": ["192.168.3.1"],
		"ReqBody": {
			"foo.com": "127.0.0.1",
			"bar.com": "192.168.4.10"
		}
	}'

	get http://foo.com/v1/foo/$(id2)
	auth testuser1
	ret 200
	json '{"foo": $(foo), "a": $(id), "b": "3"}'

	# match 是很有意思的东西，上面的 ret, json 其实都是基于 match 实现的
	# json <json-body> 等价于: match <json-body> $(resp.body)
	# ret 200 从匹配的意义来说等价于: match $(resp.code) 200
	# 但 ret 指令有表征 "发起请求并开始匹配response" 的语意，这一点无可替代
	#
	match '{"foo": $(foo2), "a": $(id), "b": "3"}' $(resp.body)
	match '["application/json"]' $(resp.header.Content-Type)
	match 200 $(resp.code)
	println \nresponse: $(resp)

	# equal 用于判断两个 json-object 是否相同
	# equal 要求变量已经绑定，且变量只能独立使用，也就是 '{"a": $(foo2)}' 这样的东西是不合法的
	#
	equal $(foo2) '"NDU3OA=="'

	# 和 equal 不同，match 并不要求两个对象完全一致
	#
	match '{"a": $(foo2)}' '{"a": "NDU3OA==", "b": 1}'

	# clear 可以用来清理一个已经绑定的变量
	#
	clear foo2
	match $(foo2) |env PATH|
	println
	println $(foo2)

	clear foo2
	match $(foo2) $(foo)
	`)

	// 下面这句和上面的 equal $(foo2) '"NDU3OA=="' 等价：
	//
	if !ctx.GetVar("foo2").Equal("NDU3OA==") {
		t.Fatal(`$(foo2) != "NDU3OA=="`)
	}
}

// ---------------------------------------------------------------------------
