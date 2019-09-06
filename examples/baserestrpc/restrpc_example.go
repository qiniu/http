package baserestrpc

import (
	"github.com/qiniu/http/httputil"
	"github.com/qiniu/http/rpcutil"
)

type fooInfo struct {
	Foo string `json:"foo"`
	A   string `json:"a"'`
	B   string `json:"b"`
	Id  string `json:"id"`
}

// ---------------------------------------------------------------------------

type Config struct {
}

type Service struct {
	foos map[string]fooInfo
}

func New(cfg *Config) (p *Service, err error) {

	p = &Service{
		foos: make(map[string]fooInfo),
	}
	return
}

// ---------------------------------------------------------------------------

type fooBarArgs struct {
	CmdArgs []string
	A       string `json:"a"'`
	B       string `json:"b"`
}

type fooBarRet struct {
	Id string `json:"id"`
}

/*
POST /foo/<FooArg>/bar
JSON {a: <A>, b: <B>}
 RET 200
JSON {id: <FooId>}
*/
func (p *Service) PostFoo_Bar(args *fooBarArgs, env *rpcutil.Env) (ret fooBarRet, err error) {

	id := args.A + "." + args.B
	p.foos[id] = fooInfo{
		Foo: args.CmdArgs[0],
		A:   args.A,
		B:   args.B,
		Id:  id,
	}
	return fooBarRet{Id: id}, nil
}

// ---------------------------------------------------------------------------

type reqArgs struct {
	CmdArgs []string
}

/*
 GET /foo/<FooId>
 RET 200
JSON {a: <A>, b: <B>, foo: <Foo>, id: <FooId>}
*/
func (p *Service) GetFoo_(args *reqArgs, env *rpcutil.Env) (ret fooInfo, err error) {

	id := args.CmdArgs[0]
	if foo, ok := p.foos[id]; ok {
		return foo, nil
	}
	err = httputil.NewError(404, "id not found")
	return
}

// ---------------------------------------------------------------------------

type postHostsArgs struct {
	CmdArgs []string
	ReqBody map[string]interface{}
}

/*
POST /hosts/<IP>
JSON {<Domain1>: <IP1>, ...}
 RET 200
JSON {"CmdArgs": [<IP>], "ReqBody": {<Domain1>: <IP1>, ...}}
*/
func (p *Service) PostHosts_(args *postHostsArgs, env *rpcutil.Env) (ret *postHostsArgs, err error) {

	return args, nil
}

// ---------------------------------------------------------------------------
