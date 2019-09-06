package authstub

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/qiniu/http/formutil"
	"github.com/qiniu/http/httputil"

	. "github.com/qiniu/http/examples/auth/proto"
)

var (
	ErrBadToken = httputil.NewError(401, "bad token")
)

// ---------------------------------------------------------------------------

func appendUint(form []byte, k string, v uint64) []byte {

	form = append(form, k...)
	return strconv.AppendUint(form, v, 10)
}

func appendString(form []byte, k, v string) []byte {

	form = append(form, k...)
	return append(form, url.QueryEscape(v)...)
}

// QiniuStub uid=$(Uid)&ut=$(Utype)&app=$(Appid)&suid=$(SudoerUid)&sut=$(SudoerUtype)
//           &ak=$(access)&eu=$(enduser_info)
//
func Format(user *SudoerInfo) string {

	return "QiniuStub " + FormatToken(user)
}

func FormatToken(user *SudoerInfo) string {

	form := make([]byte, 0, 128)
	form = appendUint(form, "uid=", uint64(user.Uid))
	form = appendUint(form, "&ut=", uint64(user.Utype))
	if user.Appid != 0 {
		form = appendUint(form, "&app=", uint64(user.Appid))
	}
	if user.Sudoer != 0 {
		form = appendUint(form, "&suid=", uint64(user.Sudoer))
		if user.UtypeSu != 0 {
			form = appendUint(form, "&sut=", uint64(user.UtypeSu))
		}
	}
	if user.Access != "" {
		form = appendString(form, "&ak=", user.Access)
	}
	if user.EndUser != "" {
		form = appendString(form, "&eu=", user.EndUser)
	}
	return string(form)
}

func Parse(auth string) (user SudoerInfo, err error) {

	if strings.HasPrefix(auth, "QiniuStub ") {
		return ParseToken(auth[10:])
	}
	err = ErrBadToken
	return
}

func ParseToken(token string) (user SudoerInfo, err error) {

	m, err := url.ParseQuery(token)
	if err != nil {
		err = ErrBadToken
		return
	}

	err = formutil.Parse(&user, m)
	if err != nil {
		err = ErrBadToken
	}
	return
}

// ---------------------------------------------------------------------------
