package restrpc

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"syscall"

	"github.com/qiniu/http/formutil"
	"github.com/qiniu/http/hfac"
	"github.com/qiniu/http/rpcutil"
)

// Env represents env of a http handler.
type Env = rpcutil.Env

/* ---------------------------------------------------------------------------

1. 参数规格

注意：以下的 [Arguments] 和 [Return-Info] 与 wsrpc 中的说明一致。

allow POST:

/foo			func (rcvr *XXXX) PostFoo([Arguments])([Return-Info])
/foo/			func (rcvr *XXXX) PostFoo_([Arguments])([Return-Info])
/foo/bar		func (rcvr *XXXX) PostFooBar([Arguments])([Return-Info])
/foo/bar/		func (rcvr *XXXX) PostFooBar_([Arguments])([Return-Info])
/foo/<cmd>/bar		func (rcvr *XXXX) PostFoo_Bar([Arguments])([Return-Info])
/foo/<cmd>/bar/<cmd>	func (rcvr *XXXX) PostFoo_Bar_([Arguments])([Return-Info])

其余的 Method(GET/PUT/DELETE) 均只需要将函数名前的前缀从 Get 改为 Get/Put/Delete 既可。

2. 参数解析

这里以 PostFoo_Bar_ 为例:

	type Args struct {
		CmdArgs []string
		FormParam1 string `json:"form_param1"`
		FormParam2 string `json:"form_param2"`
	}

	func (rcvr *XXXX) PostFoo_Bar_(args *Args, env *rpcutil.Env) {
		...
	}

如果请求为：

	POST /foo/COMMAND1/bar/COMMAND2
	Content-Type: application/x-www-form-urlencoded

	form_param1=FORM_PARAM1&form_param2=FORM_PARAM2

那么解析出来的 args 为：

	args = &Args{
		CmdArgs:    []string{"COMMAND1","COMMAND2"},
		FormParam1: "FORM_PARAM1",
		FormParam2: "FORM_PARAM2",
	}

// -------------------------------------------------------------------------*/

func isJSONCall(req *http.Request) bool {

	ct := req.Header.Get("Content-Type")
	return ct == "application/json" || strings.HasPrefix(ct, "application/json;")
}

func parseReqDefault(ret reflect.Value, req *http.Request) error {

	if cmdArgs := req.Header["*"]; cmdArgs != nil {
		v := ret.Elem().FieldByName("CmdArgs")
		if v.IsValid() {
			v.Set(reflect.ValueOf(cmdArgs))
			if ret.Elem().NumField() == 1 {
				return nil
			}
		}
	}

	if isJSONCall(req) {
		if req.ContentLength == 0 {
			return nil
		}
		return json.NewDecoder(req.Body).Decode(ret.Interface())
	}

	err := req.ParseForm()
	if err != nil {
		return err
	}
	return formutil.ParseValue(ret, req.Form, "json")
}

/* ---------------------------------------------------------------------------

在少数情况下，需用 ReqBody 来存储参数。样例：

	type Args struct {
		CmdArgs []string
		ReqBody map[string]interface{}
	}

	func (rcvr *XXXX) PostFoo_Bar_(args *Args, env *rpcutil.Env) {
		...
	}

如果请求为：

	POST /foo/COMMAND1/bar/COMMAND2
	Content-Type: application/json

	{
		"domain1": "IP1",
		"domain2": "IP2"
	}

那么解析出来的 args 为：

	args = &Args{
		CmdArgs: []string{"COMMAND1","COMMAND2"},
		ReqBody: map[string]interface{}{
			"domain1": "IP1",
			"domain2": "IP2",
		},
	}

// -------------------------------------------------------------------------*/

func parseReqWithBody(ret reflect.Value, req *http.Request) error {

	retElem := ret.Elem()
	if cmdArgs := req.Header["*"]; cmdArgs != nil {
		v := retElem.FieldByName("CmdArgs")
		if v.IsValid() {
			v.Set(reflect.ValueOf(cmdArgs))
		}
	}

	if isJSONCall(req) {
		ret = retElem.FieldByName("ReqBody").Addr()
		if req.ContentLength == 0 {
			return nil
		}
		return json.NewDecoder(req.Body).Decode(ret.Interface())
	}
	return syscall.EINVAL
}

func parseReqWithReader(ret reflect.Value, req *http.Request) error {

	retElem := ret.Elem()
	if cmdArgs := req.Header["*"]; cmdArgs != nil {
		v := retElem.FieldByName("CmdArgs")
		if v.IsValid() {
			v.Set(reflect.ValueOf(cmdArgs))
		}
	}
	retElem.FieldByName("ReqBody").Set(reflect.ValueOf(req.Body))
	return nil
}

func parseReqWithBytes(ret reflect.Value, req *http.Request) error {

	retElem := ret.Elem()
	if cmdArgs := req.Header["*"]; cmdArgs != nil {
		v := retElem.FieldByName("CmdArgs")
		if v.IsValid() {
			v.Set(reflect.ValueOf(cmdArgs))
		}
	}

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	retElem.FieldByName("ReqBody").Set(reflect.ValueOf(b))
	return nil
}

// ---------------------------------------------------------------------------

var unusedReadCloser *io.ReadCloser
var typeOfIoReadCloser = reflect.TypeOf(unusedReadCloser).Elem()

func selParseReq(reqType reflect.Type) func(ret reflect.Value, req *http.Request) error {

	if sf, ok := reqType.FieldByName("ReqBody"); ok {
		t := sf.Type
		switch t.Kind() {
		case reflect.Map:
		case reflect.Interface:
			if typeOfIoReadCloser.Implements(sf.Type) { // io.ReadCloser
				return parseReqWithReader
			}
		case reflect.Slice:
			if t.Elem().Kind() == reflect.Uint8 { // []byte
				return parseReqWithBytes
			}
		}
		return parseReqWithBody
	}
	return parseReqDefault
}

// ---------------------------------------------------------------------------

var newHandler = rpcutil.HandlerCreator{SelParseReq: selParseReq}.New

// Factory is a HandlerFactory.
var Factory = hfac.HandlerFactory{
	{"Post", newHandler},
	{"Put", newHandler},
	{"Delete", newHandler},
	{"Get", newHandler},
}

// ---------------------------------------------------------------------------
