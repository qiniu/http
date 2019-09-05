package strconv

import (
	"reflect"
	"strconv"
	"syscall"
)

// --------------------------------------------------------------------

func Parse(ret interface{}, str string) (err error) {

	v := reflect.ValueOf(ret)
	if v.Kind() != reflect.Ptr {
		return syscall.EINVAL
	}

	return ParseValue(v.Elem(), str)
}

func ParseValue(v reflect.Value, str string) (err error) {

	var iv int64
	var uv uint64
	var fv float64

retry:
	switch v.Kind() {
	case reflect.String:
		v.SetString(str)
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			v.SetBytes([]byte(str))
		} else {
			return syscall.EINVAL
		}
	case reflect.Int:
		iv, err = strconv.ParseInt(str, 10, 0)
		v.SetInt(iv)
	case reflect.Uint:
		uv, err = strconv.ParseUint(str, 10, 0)
		v.SetUint(uv)
	case reflect.Int64:
		iv, err = strconv.ParseInt(str, 10, 64)
		v.SetInt(iv)
	case reflect.Uint32:
		uv, err = strconv.ParseUint(str, 10, 32)
		v.SetUint(uv)
	case reflect.Int32:
		iv, err = strconv.ParseInt(str, 10, 32)
		v.SetInt(iv)
	case reflect.Uint16:
		uv, err = strconv.ParseUint(str, 10, 16)
		v.SetUint(uv)
	case reflect.Int16:
		iv, err = strconv.ParseInt(str, 10, 16)
		v.SetInt(iv)
	case reflect.Uint64:
		uv, err = strconv.ParseUint(str, 10, 64)
		v.SetUint(uv)
	case reflect.Ptr:
		elem := reflect.New(v.Type().Elem())
		v.Set(elem)
		v = elem.Elem()
		goto retry
	case reflect.Struct:
		method := v.Addr().MethodByName("ParseValue") // ParseValue(str string) error
		if method.IsValid() {
			out := method.Call([]reflect.Value{reflect.ValueOf(str)})
			ret := out[0].Interface()
			if ret != nil {
				return ret.(error)
			}
			return nil
		}
		return syscall.EINVAL
	case reflect.Uint8:
		uv, err = strconv.ParseUint(str, 10, 8)
		v.SetUint(uv)
	case reflect.Int8:
		iv, err = strconv.ParseInt(str, 10, 8)
		v.SetInt(iv)
	case reflect.Uintptr:
		uv, err = strconv.ParseUint(str, 10, 64)
		v.SetUint(uv)
	case reflect.Float64:
		fv, err = strconv.ParseFloat(str, 64)
		v.SetFloat(fv)
	case reflect.Float32:
		fv, err = strconv.ParseFloat(str, 32)
		v.SetFloat(fv)
	case reflect.Bool:
		var bv bool
		bv, err = strconv.ParseBool(str)
		v.SetBool(bv)
	default:
		return syscall.EINVAL
	}
	return
}

// --------------------------------------------------------------------
