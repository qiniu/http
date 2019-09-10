package restrpc

import (
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/qiniu/http/hfac"
)

// ---------------------------------------------------------------------------

type Mux interface {
	Handle(pattern string, handler http.Handler)
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	SetDefault(handler http.Handler)
}

type Router struct {
	Factory       hfac.HandlerFactory
	PatternPrefix string
	Separator     string
	Mux           Mux
	Default       http.Handler
}

func (r *Router) ListenAndServe(addr string, rcvr interface{}) error {

	return http.ListenAndServe(addr, r.Register(rcvr))
}

func (r *Router) Register(rcvr interface{}, routes ...[][2]string) Mux {

	if r.Mux == nil {
		r.Mux = NewServeMux()
	}
	if r.Default != nil {
		r.Mux.SetDefault(r.Default)
	}
	mux := r.Mux

	factory := r.Factory
	if factory == nil {
		factory = Factory
	}

	typ := reflect.TypeOf(rcvr)
	rcvr1 := reflect.ValueOf(rcvr)

	if len(routes) == 0 {
		patternPrefix := r.PatternPrefix
		if strings.HasPrefix(patternPrefix, "/") {
			patternPrefix = patternPrefix[1:]
		}
		sep := r.Separator
		if sep == "" {
			sep = "_"
		}
		// Install the methods
		for m := 0; m < typ.NumMethod(); m++ {
			method := typ.Method(m)
			prefix, handler, err := factory.Create(rcvr1, method)
			if err != nil {
				continue
			}
			pattern := []string{prefix}
			if patternPrefix != "" {
				pattern = append(pattern, patternPrefix)
			}
			pattern = append(pattern, patternOf(method.Name[len(prefix):], sep)...)
	
			mux.Handle(strings.Join(pattern, "/"), handler)
			log.Println("Install", pattern, "=>", method.Name)
		}
	} else {
		for _, item := range routes[0] {
			pattern := item[0]
			if r.PatternPrefix != "" {
				pos := strings.Index(pattern, "/")
				if pos > 0 {
					pattern = pattern[:pos] + r.PatternPrefix + pattern[pos:]
				}
			}
			method, ok := typ.MethodByName(item[1])
			if !ok {
				log.Fatalln("Install", pattern, "=>", item[1], "failed: method not found!")
			}
			_, handler, err := factory.Create(rcvr1, method)
			if err != nil {
				log.Fatalln("Install", pattern, "=>", item[1], "failed:", err)
			}
			mux.Handle(pattern, handler)
			log.Println("Install", pattern, "=>", item[1])
		}
	}
	return mux
}

//
// AppleBanana => ["Apple", "Banana"]
// Apple_Banana => ["Apple", "*", "Banana"]
// AppleBanana_ => ["Apple", "Banana", "*"]
// Apple_Banana_ => ["Apple", "*", "Banana", "*"]
// ...
//
func patternOf(method string, sep string) (pattern []string) {

	for method != "" {
		pos := strings.Index(method, sep)
		if pos == -1 {
			return appendPattern(pattern, method)
		}
		if pos > 0 {
			pattern = appendPattern(pattern, method[:pos])
		}
		pattern = append(pattern, "*")
		method = method[pos+len(sep):]
	}
	return
}

func appendPattern(pattern []string, method string) []string {

	var i, last int
	for i = 1; i < len(method); i++ {
		c := method[i]
		if c >= 'A' && c <= 'Z' {
			pattern = append(pattern, method[last:i])
			last = i
		}
	}
	return append(pattern, method[last:i])
}

// ---------------------------------------------------------------------------
