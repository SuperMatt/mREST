//Package mrest generates a gorilla mux from a specially formatted struct, to be used with net/http
package mrest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"reflect"
	"strings"

	"github.com/gorilla/mux"
)

//Request contains information about your request.
//Pass a pointer to your method functions in order to access their data.
type Request struct {
	HTTPRequest *http.Request
	Vars        map[string]string
}

//Data contains all of the information you may require for a response wrapper
type Data struct {
	Code    int
	Request *http.Request
	Data    *interface{}
}

var hlist = make(map[string]bool)
var methods = map[string]bool{"GET": true, "POST": true, "DELETE": true, "PATCH": true, "PUT": true}

func value(g interface{}, param string, r *Request) interface{} {
	if param != "" {
		return r.Vars[param]
	}
	return g
}

//GenMux will generate the mux, with a default return format
func GenMux(b string, d interface{}) *mux.Router {
	fn := func(m *Data) interface{} {
		return m.Data
	}
	return GenMuxWithResponder(b, d, fn)
}

//GenMuxWithResponder will generate a mux with a custom response wrapper
func GenMuxWithResponder(b string, d interface{}, rfn func(*Data) interface{}) *mux.Router {

	r := mux.NewRouter()

	loopMux(b, d, r, rfn)

	rootFunc := func(r *Request) (interface{}, int) { return d, http.StatusOK }
	applyMux("GET", b, rootFunc, r, rfn)

	return r
}

func loopMux(b string, d interface{}, r *mux.Router, rfn func(*Data) interface{}) {
	size := reflect.TypeOf(d).NumField()

	for i := 0; i < size; i++ {
		f := reflect.TypeOf(d).Field(i)
		v := reflect.ValueOf(d).Field(i)

		n := strings.ToLower(f.Name)

		p := path.Join(b, n)
		param, ok := f.Tag.Lookup("param")
		if ok {
			p = path.Join(p, "{"+param+"}")
		}

		if v.CanInterface() {
			numMethods := v.NumMethod()
			t := reflect.TypeOf(v.Interface())
			if numMethods > 0 {
				for j := 0; j < numMethods; j++ {
					m := t.Method(j).Name
					_, ok := methods[m]
					if ok {
						mi := v.MethodByName(m).Interface().(func(*Request) (interface{}, int))
						applyMux(m, p, mi, r, rfn)

					}
				}
			}

			g := v.Interface()
			if v.Kind() == reflect.Struct {
				loopMux(p, g, r, rfn)
			}

			fn := func(r *Request) (interface{}, int) { return value(g, param, r), http.StatusOK }
			applyMux("GET", p, fn, r, rfn)
		}
	}
}

func applyMux(m string, path string, fn func(*Request) (interface{}, int), r *mux.Router, rfn func(*Data) interface{}) {
	key := m + path
	_, ok := hlist[key]
	if !ok {
		hlist[key] = true
		hf := func(w http.ResponseWriter, req *http.Request) {
			d := Request{Vars: mux.Vars(req), HTTPRequest: req}
			output, code := fn(&d)
			meta := Data{Code: code, Request: req, Data: &output}
			resp := rfn(&meta)

			o, err := json.Marshal(resp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, err.Error())
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)
			fmt.Fprintf(w, string(o))
		}

		r.HandleFunc(path, hf).Methods(m)
		r.HandleFunc(path+"/", hf).Methods(m)
	}
}
