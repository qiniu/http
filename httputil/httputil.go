package httputil

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"syscall"
)

// ---------------------------------------------------------------------------
// type ErrorInfo

// ErrorInfo represents a rpc error.
type ErrorInfo struct {
	Err   string `json:"error,omitempty"`
	Errno int    `json:"errno,omitempty"`
	Code  int    `json:"code"`
}

// NewError creates a rpc error.
func NewError(code int, err string) *ErrorInfo {

	return &ErrorInfo{Code: code, Err: err}
}

// NewErrorEx creates a rpc error.
func NewErrorEx(code, errno int, err string) *ErrorInfo {

	return &ErrorInfo{Code: code, Errno: errno, Err: err}
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

// StatusCode gets http status code.
func (r *ErrorInfo) StatusCode() int {

	return r.Code
}

// ---------------------------------------------------------------------------

type nestedObjectGetter interface {
	NestedObject() interface{}
}

// GetErrorInfo returns http status code and an error message.
func GetErrorInfo(err error) (code, errno int, errmsg string) {

	if e, ok := err.(*ErrorInfo); ok {
		return e.Code, e.Errno, e.Error()
	}
	if getter, ok := err.(nestedObjectGetter); ok {
		if e, ok := getter.NestedObject().(*ErrorInfo); ok {
			return e.Code, e.Errno, e.Error()
		}
	}
	switch err {
	case syscall.EINVAL:
		return 400, 0, "invalid arguments"
	case syscall.ENOENT: // no such entry
		return 404, 0, "entry not found"
	case syscall.EEXIST: // entry exists
		return 409, 0, "entry already exists"
	}
	return 500, 0, err.Error()
}

// ---------------------------------------------------------------------------

type errorRet struct {
	Err   string `json:"error"`
	Errno int    `json:"errno,omitempty"`
}

// Error replies an error as a http response.
func Error(w http.ResponseWriter, err error) {

	if err == nil {
		h := w.Header()
		h.Set("Content-Length", "2")
		h.Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(emptyObj)
		return
	}

	code, errno, errmsg := GetErrorInfo(err)
	Reply(w, code, &errorRet{Err: errmsg, Errno: errno})
}

// ReplyErr replies an error as a http response.
func ReplyErr(w http.ResponseWriter, code int, err string) {

	Reply(w, code, &errorRet{Err: err})
}

// ---------------------------------------------------------------------------
// func Reply

// Reply replies a http response.
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

// ReplyWith replies a http response.
func ReplyWith(w http.ResponseWriter, code int, bodyType string, msg []byte) {

	h := w.Header()
	h.Set("Content-Length", strconv.Itoa(len(msg)))
	h.Set("Content-Type", bodyType)
	w.WriteHeader(code)
	w.Write(msg)
}

// ReplyWithStream replies a http response.
func ReplyWithStream(w http.ResponseWriter, code int, bodyType string, body io.Reader, bytes int64) {

	h := w.Header()
	h.Set("Content-Length", strconv.FormatInt(bytes, 10))
	h.Set("Content-Type", bodyType)
	w.WriteHeader(code)
	io.Copy(w, body) // don't use io.CopyN: if you need, call io.LimitReader(body, bytes) by yourself
}

// ReplyWithCode replies a http response.
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
