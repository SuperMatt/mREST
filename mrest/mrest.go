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

//Request is
type Request struct {
	HTTPRequest *http.Request
	Vars        map[string]string
}

var hlist = make(map[string]bool)

var methods = map[string]bool{"GET": true, "POST": true, "DELETE": true, "PATCH": true, "PUT": true}

//RestServer is the main configuration block for our restful apis
type RestServer struct {
	server    *http.Server
	dummyData *interface{}
}

//NewRest creates a simple RestServer pointer.
func NewRest(addr string, dummydata *interface{}) (r *RestServer, err error) {
	h := &http.Server{Addr: addr}
	r, err = NewRestWithServer(h, dummydata)
	return r, err
}

//NewRestWithServer creates a server based on a pointer to an existing http.Server instance, allowing greater control
func NewRestWithServer(h *http.Server, dummydata *interface{}) (r *RestServer, err error) {
	var rs RestServer
	rs.setServer(h)
	rs.setDummyData(dummydata)
	return &rs, nil
}

func (rs RestServer) setServer(h *http.Server) {
	rs.server = h
}

func (rs RestServer) setDummyData(d *interface{}) {
	rs.dummyData = d
}

func value(g interface{}, param string, r *Request) interface{} {
	if param != "" {
		return r.Vars[param]
	}
	return g
}

//GenMux will generate the mux, with a default return format
func GenMux(b string, d interface{}) *mux.Router {
	fn := func(c int, i interface{}, r *http.Request) interface{} {
		return i
	}
	return GenMuxWithResponder(b, d, fn)
}

//GenMuxWithResponder will generate a mux where the user can set the default format
func GenMuxWithResponder(b string, d interface{}, rfn func(int, interface{}, *http.Request) interface{}) *mux.Router {

	r := mux.NewRouter()

	loopMux(b, d, r, rfn)

	rootFunc := func(r *Request) (interface{}, int) { return d, http.StatusOK }
	applyMux("GET", b, rootFunc, r, rfn)

	return r
}

func loopMux(b string, d interface{}, r *mux.Router, rfn func(int, interface{}, *http.Request) interface{}) {
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

func applyMux(m string, path string, fn func(*Request) (interface{}, int), r *mux.Router, rfn func(int, interface{}, *http.Request) interface{}) {
	key := m + path
	_, ok := hlist[key]
	if !ok {
		hlist[key] = true
		hf := func(w http.ResponseWriter, req *http.Request) {
			d := Request{Vars: mux.Vars(req), HTTPRequest: req}
			output, code := fn(&d)
			resp := rfn(code, output, req)

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
