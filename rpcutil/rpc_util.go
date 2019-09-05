package rpcutil

import (
	"context"
	"io"
	"log"
	"net/http"
	"reflect"
	"syscall"

	"github.com/qiniu/http/httputil"
)

// ---------------------------------------------------------------------------

// Env represents env of a http handler.
type Env struct {
	W   http.ResponseWriter
	Req *http.Request
}

// ---------------------------------------------------------------------------

var unusedRespW *http.ResponseWriter
var unusedReq *http.Request
var unusedContext *context.Context
var typeOfRespW = reflect.TypeOf(unusedRespW).Elem()
var typeOfReq = reflect.TypeOf(unusedReq)
var typeOfContext = reflect.TypeOf(unusedContext).Elem()

func isEnv(t reflect.Type) bool {
	return t.NumField() == 2 &&
		t.Field(0).Type == typeOfRespW &&
		t.Field(1).Type == typeOfReq
}

func setEnv(v reflect.Value, w http.ResponseWriter, req *http.Request) {
	v.Field(0).Set(reflect.ValueOf(w))
	v.Field(1).Set(reflect.ValueOf(req))
}

// ---------------------------------------------------------------------------

type itfEnv interface {
	OpenEnv(rcvr interface{}, w *http.ResponseWriter, req *http.Request) error
	CloseEnv()
}

// ---------------------------------------------------------------------------

// Replier represents a http replier.
type Replier struct {
	Reply         func(w http.ResponseWriter, code int, data interface{})
	ReplyWithCode func(w http.ResponseWriter, code int)
	Error         func(w http.ResponseWriter, err error)
}

var defaultRepl = &Replier{
	Reply:         httputil.Reply,
	ReplyWithCode: httputil.ReplyWithCode,
	Error:         httputil.Error,
}

/* ---------------------------------------------------------------------------

func (rcvr *XXXX) YYYY(req ZZZZ, env ENV) (err error)
func (rcvr *XXXX) YYYY(req ZZZZ, env ENV) (ret RRRR, err error)
func (rcvr *XXXX) YYYY(req ZZZZ, env ENV)

func (rcvr *XXXX) YYYY(req ZZZZ) (err error)
func (rcvr *XXXX) YYYY(req ZZZZ) (ret RRRR, err error)

func (rcvr *XXXX) YYYY(env ENV) (err error)
func (rcvr *XXXX) YYYY(env ENV) (ret RRRR, err error)

func (rcvr *XXXX) YYYY() (err error)
func (rcvr *XXXX) YYYY() (ret RRRR, err error)

在[]里面的参数是可选的(因为太多了，没有把Context放到上面的列表里面)
func (rcvr *XXXX) YYYY(ctx Context[, req ZZZZ][, env ENV]) ([ret RRRR, ]err error)

// -------------------------------------------------------------------------*/

var errMustPOST = httputil.NewError(http.StatusMethodNotAllowed, "Request method must be POST")

type readerCloser struct {
	io.Reader
	io.Closer
}

type handler struct {
	rcvr      reflect.Value
	method    reflect.Value
	reqType   reflect.Type
	envType   reflect.Type
	parseReq  func(v reflect.Value, req *http.Request) error
	repl      *Replier
	hasEnv    int16 // 0: no env  1: Env  2: IEnv  3: deprecated Env  0x83: deprecated *Env
	hasRet    int8  // -1: no ret  0: (err error)  1: (ret RRRR, err error)
	hasCtx    int8  // 0: no Context 1: has Context
	reqNotPtr int16
	postOnly  int16
}

var zero reflect.Value

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	if h.postOnly == 1 && req.Method != "POST" {
		httputil.Error(w, errMustPOST)
		return
	}

	repl := h.repl

	var typeAddr, ctxAddr *reflect.Value

	args := make([]reflect.Value, 0, 4)

	args = append(args, h.rcvr)

	if h.hasCtx == 1 {
		args = append(args, zero)
		ctxAddr = &args[len(args)-1]
	}

	if h.reqType != nil {
		args = append(args, zero)
		typeAddr = &args[len(args)-1]
	}

	var err error
	switch h.hasEnv {
	case 0:
	case 1:
		args = append(args, reflect.ValueOf(Env{w, req}))
	case 2:
		env := reflect.New(h.envType)
		env1 := env.Interface().(itfEnv)
		err = env1.OpenEnv(h.rcvr.Interface(), &w, req)
		if err != nil {
			repl.Error(w, err)
			return
		}
		defer env1.CloseEnv()
		args = append(args, env)
	case 3:
		env := reflect.New(h.envType).Elem()
		setEnv(env, w, req)
		args = append(args, env)
	case 0x83:
		env := reflect.New(h.envType)
		setEnv(env.Elem(), w, req)
		args = append(args, env)
	}

	if h.reqType != nil {
		req1 := reflect.New(h.reqType)
		err = h.parseReq(req1, req)
		if err != nil {
			err2 := httputil.NewError(400, err.Error())
			repl.Error(w, err2)
			return
		}
		if h.reqNotPtr != 0 {
			req1 = req1.Elem()
		}
		*typeAddr = req1
	}

	var out []reflect.Value
	if h.hasCtx == 1 {
		*ctxAddr = reflect.ValueOf(context.Background())
	}

	out = h.method.Call(args)
	if h.hasRet < 0 {
		return
	}

	err1 := out[h.hasRet]
	if !err1.IsNil() {
		e := err1.Interface().(error)
		repl.Error(w, e)
		return
	}

	if h.hasRet != 0 {
		repl.Reply(w, 200, out[0].Interface())
	} else {
		repl.ReplyWithCode(w, 200)
	}
}

// ---------------------------------------------------------------------------

// Precompute the reflect type for `error`. Can't use `error` directly
// because Typeof takes an empty interface value. This is annoying.
var unusedError *error
var unusedEnv *Env
var unusedIEnv *itfEnv
var typeOfError = reflect.TypeOf(unusedError).Elem()
var typeOfEnv = reflect.TypeOf(unusedEnv).Elem()
var typeOfIEnv = reflect.TypeOf(unusedIEnv).Elem()

// HandlerCreator represents a http handler creator.
type HandlerCreator struct {
	ParseReq     func(v reflect.Value, req *http.Request) error
	SelParseReq  func(reqType reflect.Type) func(v reflect.Value, req *http.Request) error
	Repl         *Replier
	ReqMayNotPtr bool
	PostOnly     bool
}

// New creates a http handler.
func (p HandlerCreator) New(rcvr reflect.Value, method reflect.Method) (http.Handler, error) {

	mtype := method.Type

	//
	// Method spec:
	//  (rcvr *XXXX) YYYY(req ZZZZ, env ENV) (err error)
	//  (rcvr *XXXX) YYYY(req ZZZZ, env ENV) (ret RRRR, err error)
	//  (rcvr *XXXX) YYYY(req ZZZZ, env ENV)
	//  (rcvr *XXXX) YYYY(req ZZZZ) (ret RRRR, err error)
	//  (rcvr *XXXX) YYYY(req ZZZZ) (err error)
	//  (rcvr *XXXX) YYYY(env ENV) (err error)
	//  (rcvr *XXXX) YYYY(env ENV) (ret RRRR, err error)
	//  (rcvr *XXXX) YYYY() (err error)
	//  (rcvr *XXXX) YYYY() (ret RRRR, err error)
	//
	// 在[]里面的参数是可选的(因为太多了，没有把Context放到上面的列表里面)
	//  (rcvr *XXXX) YYYY(ctx Context[, req ZZZZ][, env ENV]) ([ret RRRR, ]err error)
	//
	var envType reflect.Type
	var hasEnv = 0
	var hasRet = mtype.NumOut() - 1
	var hasCtx = 0
	var narg = mtype.NumIn()
	if narg > 1 {
		lastArg := mtype.In(narg - 1)
		if lastArg.Kind() == reflect.Struct {
			if lastArg == typeOfEnv {
				hasEnv = 1
				narg--
			} else if isEnv(lastArg) { // deprecated Env
				envType = lastArg
				hasEnv = 3
				narg--
			}
		} else if lastArg.Kind() == reflect.Ptr {
			if lastArg.Implements(typeOfIEnv) {
				envType = lastArg.Elem()
				hasEnv = 2
				narg--
			} else {
				lastArg = lastArg.Elem()
				if isEnv(lastArg) { // deprecated *Env
					envType = lastArg
					hasEnv = 0x83
					narg--
				}
			}
		}
	}

	var nargBase = 0
	if narg >= 2 {
		secArg := mtype.In(1)

		if secArg.Implements(typeOfContext) {
			hasCtx = 1
			narg--
			nargBase++
		}
	}

	if (hasRet < -1 || hasRet > 2) || (narg != 2 && narg != 1) {
		log.Println(
			"method", method.Name,
			"has wrong number arguments or return values:", mtype.NumIn(), mtype.NumOut())
		return nil, syscall.EINVAL
	}

	var reqNotPtr int16
	var reqType reflect.Type
	if narg == 2 {
		reqType = mtype.In(nargBase + 1)
		if reqType.Kind() == reflect.Ptr {
			reqType = reqType.Elem()
		} else if p.ReqMayNotPtr {
			reqNotPtr = 1
		} else {
			log.Println("method", method.Name, "arg type not a pointer:", reqType)
			return nil, syscall.EINVAL
		}
	}

	if hasRet >= 0 {
		if errType := mtype.Out(hasRet); errType != typeOfError {
			log.Println("method", method.Name, "returns", errType.String(), "not error")
			return nil, syscall.EINVAL
		}
	}

	h := &handler{
		rcvr, method.Func, reqType, envType,
		p.ParseReq, defaultRepl, int16(hasEnv), int8(hasRet), int8(hasCtx), reqNotPtr, 0}

	if h.parseReq == nil && p.SelParseReq != nil {
		if reqType != nil {
			h.parseReq = p.SelParseReq(reqType)
		}
	}

	if p.Repl != nil {
		h.repl = p.Repl
	}
	if p.PostOnly {
		h.postOnly = 1
	}
	return h, nil
}

// ---------------------------------------------------------------------------
