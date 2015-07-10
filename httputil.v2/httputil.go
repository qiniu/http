package httputil

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"syscall"

	"qiniupkg.com/x/errors.v7"
	"qiniupkg.com/x/log.v7"
)

// ---------------------------------------------------------------------------
// type ErrorInfo

type ErrorInfo struct {
	Err   string `json:"error,omitempty"`
	Key   string `json:"key,omitempty"`
	Errno int    `json:"errno,omitempty"`
	Code  int    `json:"code"`
}

func NewError(code int, err string) *ErrorInfo {

	return &ErrorInfo{Code: code, Err: err}
}

func NewRpcError(code, errno int, key, err string) *ErrorInfo {

	return &ErrorInfo{Code: code, Errno: errno, Key: key, Err: err}
}

func (r *ErrorInfo) ErrorDetail() string {

	msg, _ := json.Marshal(r)
	return string(msg)
}

func (r *ErrorInfo) Error() string {

	if r.Err != "" {
		return r.Err
	}
	if err := http.StatusText(r.Code); err != "" {
		return err
	}
	return "E" + strconv.Itoa(r.Code)
}

func (e *ErrorInfo) RpcError() (code, errno int, key, err string) {

	return e.Code, e.Errno, e.Key, e.Error()
}

func (r *ErrorInfo) HttpCode() int {

	return r.Code
}

// ---------------------------------------------------------------------------

type httpCoder interface {
	HttpCode() int
}

type nestedObjectGetter interface {
	NestedObject() interface{}
}

func DetectCode(err error) int {

	if e, ok := err.(httpCoder); ok {
		return e.HttpCode()
	}
	if getter, ok := err.(nestedObjectGetter); ok {
		if e, ok := getter.NestedObject().(httpCoder); ok {
			return e.HttpCode()
		}
	}
	switch err {
	case syscall.EINVAL:
		return 400
	case syscall.ENOENT: // no such entry
		return 612
	case syscall.EEXIST: // entry exists
		return 614
	}
	return 599
}

// ---------------------------------------------------------------------------

type rpcError interface {
	RpcError() (code, errno int, key, err string)
}

type errorRet struct {
	Err   string `json:"error"`
	Key   string `json:"key,omitempty"`
	Errno int    `json:"errno,omitempty"`
}

func replyErr(skip int, w http.ResponseWriter, err error) {

	if err == nil {
		h := w.Header()
		h.Set("Content-Length", "2")
		h.Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(emptyObj)
		return
	}

	var code int
	var ret errorRet

	if e, ok := err.(rpcError); ok {
		code, ret.Errno, ret.Key, ret.Err = e.RpcError()
	} else if getter, ok := err.(nestedObjectGetter); ok {
		if e2, ok := getter.NestedObject().(rpcError); ok {
			code, ret.Errno, ret.Key, ret.Err = e2.RpcError()
		}
	}

	if code == 0 {
		switch err {
		case syscall.EINVAL:
			code, ret.Err = 400, "invalid argument"
		case syscall.ENOENT:
			code, ret.Err = 612, "no such entry"
		case syscall.EEXIST:
			code, ret.Err = 614, "entry exists"
		default:
			code, ret.Err = 599, err.Error()
		}
	}

	detail := errors.Detail(err)
	logWithReqid(skip+1, w.Header().Get("X-Reqid"), detail)

	Reply(w, code, &ret)
}

func logWithReqid(lvl int, reqid string, str string) {

	str = strings.Replace(str, "\n", "\n["+reqid+"]", -1)
	log.Std.Output(reqid, log.Lwarn, lvl+1, str)
}

func Error(w http.ResponseWriter, err error) {

	replyErr(2, w, err)
}

func ReplyErr(w http.ResponseWriter, code int, err string) {

	replyErr(2, w, NewError(code, err))
}

// ---------------------------------------------------------------------------
// func Reply

func Reply(w http.ResponseWriter, code int, data interface{}) {

	msg, err := json.Marshal(data)
	if err != nil {
		Error(w, err)
		return
	}

	h := w.Header()
	h.Set("Content-Length", strconv.Itoa(len(msg)))
	h.Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(msg)
}

func ReplyWith(w http.ResponseWriter, code int, bodyType string, msg []byte) {

	h := w.Header()
	h.Set("Content-Length", strconv.Itoa(len(msg)))
	h.Set("Content-Type", bodyType)
	w.WriteHeader(code)
	w.Write(msg)
}

func ReplyWithStream(w http.ResponseWriter, code int, bodyType string, body io.Reader, bytes int64) {

	h := w.Header()
	h.Set("Content-Length", strconv.FormatInt(bytes, 10))
	h.Set("Content-Type", bodyType)
	w.WriteHeader(code)
	io.Copy(w, body) // don't use io.CopyN: if you need, call io.LimitReader(body, bytes) by yourself
}

func ReplyWithCode(w http.ResponseWriter, code int) {

	if code < 400 {
		h := w.Header()
		h.Set("Content-Length", "2")
		h.Set("Content-Type", "application/json")
		w.WriteHeader(code)
		w.Write(emptyObj)
	} else {
		err := http.StatusText(code)
		if err == "" {
			err = "E" + strconv.Itoa(code)
		}
		ReplyErr(w, code, err)
	}
}

var emptyObj = []byte{'{', '}'}

// ---------------------------------------------------------------------------

