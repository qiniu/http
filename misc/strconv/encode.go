package strconv

import (
	"errors"
	"reflect"
	"strconv"

	. "github.com/qiniu/http/misc/types"
)

var (
	ErrOmit          = errors.New("omit")
	ErrUnsupportType = errors.New("unsupport type")
)

func Encode(v interface{}, omitempty bool) (ret string, err error) {

	return EncodeValue(reflect.ValueOf(v), omitempty)
}

func EncodeValue(v reflect.Value, omitempty bool) (ret string, err error) {

	kind := v.Kind()
	switch {
	case kind == reflect.String:
		ret = v.String()
		if omitempty && ret == "" {
			return "", ErrOmit
		}

	case Is(kind, Ints):
		val := v.Int()
		if omitempty && val == 0 {
			return "", ErrOmit
		}
		ret = strconv.FormatInt(val, 10)

	case Is(kind, Uints):
		val := v.Uint()
		if omitempty && val == 0 {
			return "", ErrOmit
		}
		ret = strconv.FormatUint(val, 10)

	case kind == reflect.Bool:
		if v.Bool() {
			return "true", nil
		}
		if omitempty {
			return "", ErrOmit
		}
		ret = "false"

	case Is(kind, Floats):
		val := v.Float()
		if omitempty && val == 0 {
			return "", ErrOmit
		}
		bitSize := 64
		if kind == reflect.Float32 {
			bitSize = 32
		}
		ret = strconv.FormatFloat(val, 'g', -1, bitSize)

	default:
		return "", ErrUnsupportType
	}
	return
}
