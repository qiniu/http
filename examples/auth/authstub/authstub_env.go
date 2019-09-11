package authstub

import (
	"net/http"

	"github.com/qiniu/http/restrpc"
	. "github.com/qiniu/http/examples/auth/proto"
)

// ---------------------------------------------------------------------------

type Env struct {
	restrpc.Env
	UserInfo
}

func (p *Env) OpenEnv(rcvr interface{}, w *http.ResponseWriter, req *http.Request) (err error) {

	auth := req.Header.Get("Authorization")
	user, err := Parse(auth)
	if err != nil {
		return
	}

	p.Env.OpenEnv(rcvr, w, req)
	p.UserInfo = user.UserInfo
	return nil
}

func (p *Env) CloseEnv() {
}

// ---------------------------------------------------------------------------
