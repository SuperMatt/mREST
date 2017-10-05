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

//GenMux generates the handlers.
func GenMux(b string, d interface{}) *mux.Router {

	r := mux.NewRouter()

	loopMux(b, d, r)

	rootFunc := func(r *Request) interface{} { return d }
	applyMux("GET", b, rootFunc, r)

	return r
}

func loopMux(b string, d interface{}, r *mux.Router) {
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
						mi := v.MethodByName(m).Interface().(func(*Request) interface{})
						applyMux(m, p, mi, r)

					}
				}
			}

			g := v.Interface()
			if v.Kind() == reflect.Struct {
				loopMux(p, g, r)
			}

			fn := func(r *Request) interface{} { return value(g, param, r) }
			applyMux("GET", p, fn, r)
		}
	}
}

func applyMux(m string, path string, fn func(*Request) interface{}, r *mux.Router) {
	key := m + path
	_, ok := hlist[key]
	if !ok {
		hlist[key] = true
		hf := func(w http.ResponseWriter, req *http.Request) {
			d := Request{Vars: mux.Vars(req), HTTPRequest: req}
			output := fn(&d)

			o, err := json.Marshal(output)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, err.Error())
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, string(o))
		}

		r.HandleFunc(path, hf).Methods(m)
		r.HandleFunc(path+"/", hf).Methods(m)
	}
}
