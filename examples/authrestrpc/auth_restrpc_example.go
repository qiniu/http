package authrestrpc

import (
	"strconv"

	"github.com/qiniu/http/examples/auth/authstub"
	"github.com/qiniu/http/httputil"
)

type fooInfo struct {
	Foo string `json:"foo"`
	A   string `json:"a"`
	B   string `json:"b"`
	ID  string `json:"id"`
	Uid uint32 `json:"uid"`
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
	A string `json:"a"`
	B string `json:"b"`
}

type fooBarRet struct {
	ID string `json:"id"`
}

/*
PostFoo_Bar protocol:
	POST /foo/<FooArg>/bar
	JSON {a: <A>, b: <B>}
	RET 200
	JSON {id: <FooId>}
*/
func (p *Service) PostFoo_Bar(args *fooBarArgs, env *authstub.Env) (ret fooBarRet, err error) {

	id := strconv.Itoa(int(env.Uid)) + "." + args.A + "." + args.B
	p.foos[id] = fooInfo{
		Foo: env.CmdArgs[0],
		A:   args.A,
		B:   args.B,
		ID:  id,
		Uid: env.Uid,
	}
	return fooBarRet{ID: id}, nil
}

// ---------------------------------------------------------------------------

/*
GetFoo_ protocol:
	GET /foo/<FooId>
	RET 200
	JSON {a: <A>, b: <B>, foo: <Foo>, id: <FooId>}
*/
func (p *Service) GetFoo_(env *authstub.Env) (ret fooInfo, err error) {

	id := env.CmdArgs[0]
	if foo, ok := p.foos[id]; ok && foo.Uid == env.Uid {
		return foo, nil
	}
	err = httputil.NewError(404, "id not found")
	return
}

// ---------------------------------------------------------------------------

type postHostsArgs struct {
	ReqBody map[string]interface{}
}

type postHostsRet struct {
	CmdArgs []string
	ReqBody map[string]interface{}
}

/*
PostHosts_ protocol:
	POST /hosts/<IP>
	JSON {<Domain1>: <IP1>, ...}
	RET 200
	JSON {"CmdArgs": [<IP>], "ReqBody": {<Domain1>: <IP1>, ...}}
*/
func (p *Service) PostHosts_(args *postHostsArgs, env *authstub.Env) (ret postHostsRet, err error) {

	return postHostsRet{env.CmdArgs, args.ReqBody}, nil
}

// ---------------------------------------------------------------------------
