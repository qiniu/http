package formutil

import (
	"bytes"
	"errors"
	"net/url"
	"reflect"
	"strings"
	"syscall"

	"github.com/qiniu/http/misc/strconv"
)

// --------------------------------------------------------------------

// EncodeToString encodes the values into ``URL encoded'' form
// ("bar=baz&foo=quux")
func EncodeToString(v interface{}) (ret string, err error) {
	ret1, err := EncodeValue(reflect.ValueOf(v), "json")
	if err != nil {
		return
	}
	return string(ret1), nil
}

// Encode encodes a value into ``URL encoded'' form
func Encode(v interface{}) (ret []byte, err error) {
	return EncodeValue(reflect.ValueOf(v), "json")
}

// EncodeValue encodes a value into ``URL encoded'' form
func EncodeValue(v reflect.Value, cate string) (ret []byte, err error) {

retry:
	switch v.Kind() {
	case reflect.Struct:
		return encodeStructValue(v, cate)
	case reflect.Ptr:
		v = v.Elem()
		goto retry
	default:
		return nil, syscall.EINVAL
	}
}

// --------------------------------------------------------------------

func encodeStructValue(v reflect.Value, cate string) (ret []byte, err error) {

	var buf bytes.Buffer

	vt := v.Type()
	n := vt.NumField()
	for i := 0; i < n; i++ {
		sf := vt.Field(i)
		if sf.Tag == "" { // no tag, skip
			continue
		}
		tag, opts, err2 := parseEncodeTag(sf.Tag.Get(cate))
		if err2 != nil {
			return nil, err2
		}

		err2 = encodeValue(&buf, tag+"=", v.Field(i), opts.omitempty)
		if err2 != nil && err2 != strconv.ErrOmit {
			return nil, err2
		}
	}
	return buf.Bytes(), nil
}

func encodeValue(buf *bytes.Buffer, prefix string, v reflect.Value, omitempty bool) (err error) {

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		n := v.Len()
		for i := 0; i < n; i++ {
			err = encodeOne(buf, prefix, v.Index(i), false)
			if err != nil {
				return
			}
		}
		return nil

	default:
		return encodeOne(buf, prefix, v, omitempty)
	}
}

func encodeOne(buf *bytes.Buffer, prefix string, val reflect.Value, omitempty bool) (err error) {

	v, err := strconv.EncodeValue(val, omitempty)
	if err != nil {
		return
	}

	if buf.Len() > 0 {
		buf.WriteByte('&')
	}
	buf.WriteString(prefix)
	buf.WriteString(url.QueryEscape(v))
	return nil
}

// --------------------------------------------------------------------

type tagEncodeOpts struct {
	omitempty bool
}

func parseEncodeTag(tag1 string) (tag string, opts tagEncodeOpts, err error) {

	if tag1 == "" {
		err = errors.New("Struct field has no tag")
		return
	}

	parts := strings.Split(tag1, ",")
	tag = parts[0]
	for i := 1; i < len(parts); i++ {
		switch parts[i] {
		case "omitempty":
			opts.omitempty = true
		default:
			err = errors.New("Unknown tag option: " + parts[i])
			return
		}
	}
	return
}

// --------------------------------------------------------------------
