package hfac

import (
	"errors"
	"log"
	"net/http"
	"reflect"
	"syscall"

	"github.com/qiniu/http/hfac/ctype"
)

/* ---------------------------------------------------------------------------

func (rcvr *XXXX) DoYYYY(w http.ResponseWriter, req *http.Request)

// -------------------------------------------------------------------------*/

type handler struct {
	rcvr   reflect.Value
	method reflect.Value
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	w1 := reflect.ValueOf(w)
	req1 := reflect.ValueOf(req)
	h.method.Call([]reflect.Value{h.rcvr, w1, req1})
}

// ---------------------------------------------------------------------------

// Precompute the reflect type for http.ResponseWriter. Can't use http.ResponseWriter directly
// because Typeof takes an empty interface value. This is annoying.
var unusedResponseWriter *http.ResponseWriter
var unusedRequest *http.Request

var typeOfHttpResponseWriter = reflect.TypeOf(unusedResponseWriter).Elem()
var typeOfHttpRequest = reflect.TypeOf(unusedRequest)

func NewHandler(rcvr reflect.Value, method reflect.Method) (http.Handler, error) {

	mtype := method.Type

	// Method spec:
	//  (rcvr *XXXX) DoYYYY(w http.ResponseWriter, req *http.Request)
	if mtype.NumOut() != 0 || mtype.NumIn() != 3 {
		log.Println("method", method.Name, "has wrong number arguments or return values:", mtype.NumIn(), mtype.NumOut())
		return nil, syscall.EINVAL
	}

	// First arg muste be http.ResponseWriter
	if wType := mtype.In(1); wType != typeOfHttpResponseWriter {
		log.Println("method", method.Name, "first argument type not http.ResponseWriter:", wType)
		return nil, syscall.EINVAL
	}

	// Second arg must be *http.Request
	if reqType := mtype.In(2); reqType != typeOfHttpRequest {
		log.Println("method", method.Name, "second arguement type not *http.Request:", reqType)
		return nil, syscall.EINVAL
	}

	return &handler{rcvr, method.Func}, nil
}

// ---------------------------------------------------------------------------

type HandlerCreator struct {
	Prefix  string
	Creator func(rcvr reflect.Value, method reflect.Method) (http.Handler, error)
}

type HandlerFactory []HandlerCreator

var ErrMethodPrefix = errors.New("invalid method name prefix")

func (r HandlerFactory) Union(r2 HandlerFactory) HandlerFactory {

	ret := make(HandlerFactory, 0, len(r)+len(r2))
	r = append(ret, r...)

	n := len(r) - 1
	if n >= 0 && r[n].Prefix == "Do" {
		r = r[:n]
	}
	return append(r, r2...)
}

func (r HandlerFactory) Create(rcvr reflect.Value, method reflect.Method) (string, http.Handler, error) {

	prefix, ok := prefixOf(method.Name)
	if !ok {
		return "", nil, ErrMethodPrefix
	}

	for _, item := range r {
		if item.Prefix == prefix {
			h, err := item.Creator(rcvr, method)
			if err == nil {
				return prefix, h, nil
			}
			return "", nil, err
		}
	}
	return "", nil, ErrMethodPrefix
}

func prefixOf(name string) (prefix string, ok bool) {

	if !ctype.Is(ctype.UPPER, rune(name[0])) {
		return
	}
	for i := 1; i < len(name); i++ {
		if !ctype.Is(ctype.LOWER, rune(name[i])) {
			return name[:i], true
		}
	}
	return
}

// Factory is a HandlerFactory.
var Factory = HandlerFactory{
	{Prefix: "Do", Creator: NewHandler},
}

// ---------------------------------------------------------------------------
