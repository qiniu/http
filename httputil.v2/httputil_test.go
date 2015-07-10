package httputil

import (
	"net/http/httptest"
	"syscall"
	"testing"

	"qiniupkg.com/x/errors.v7"
)

func MysqlError(err error, cmd ...interface{}) error {

	return errors.InfoEx(2, syscall.EINVAL, cmd...).Detail(err)
}

func (r *ErrorInfo) makeError() error {

	err := errors.New("detail error")
	return MysqlError(err, "do sth failed")
}

func TestError(t *testing.T) {

	err := new(ErrorInfo).makeError()
	w := httptest.NewRecorder()
	w.Header().Set("X-Reqid", "123456")
	Error(w, err)
}

