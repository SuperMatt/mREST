package mrest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"reflect"
)

type handlers map[string]bool

var hlist = make(handlers)

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

//GenMux generates the handlers.
func GenMux(b string, d interface{}) {

	loopMux(b, d)
	applyMux(b, func(r *http.Request) interface{} { return d })
}

func loopMux(b string, d interface{}) {
	size := reflect.TypeOf(d).NumField()

	for i := 0; i < size; i++ {
		f := reflect.TypeOf(d).Field(i)
		v := reflect.ValueOf(d).Field(i)

		n := f.Name

		p := path.Join(b, n)

		if v.CanInterface() {
			numMethods := v.NumMethod()
			t := reflect.TypeOf(v.Interface())
			if numMethods > 0 {
				for j := 0; j < numMethods; j++ {
					if t.Method(j).Name == "GETHandler" {
						mi := v.MethodByName("Handler").Interface().(func(*http.Request) interface{})
						applyMux(p, mi)
						applyMux(p+"/", mi)
					}
				}
			}

			g := v.Interface()
			if v.Kind() == reflect.Struct {
				loopMux(p, g)
			}
			applyMux(p, func(r *http.Request) interface{} { return g })
			applyMux(p+"/", func(r *http.Request) interface{} { return g })
		}
	}
}

func applyMux(path string, fn func(*http.Request) interface{}) {
	_, ok := hlist[path]
	if !ok {
		hlist[path] = true
		hf := func(w http.ResponseWriter, r *http.Request) {
			resp := fn(r)

			o, err := json.Marshal(resp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, err.Error())
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, string(o))
		}

		http.HandleFunc(path, hf)
	}
}
