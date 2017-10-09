package mrest

import (
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"
)

type StructTest struct {
	Apple  apple `param:"apple"`
	Banana int
	Carrot string `param:"carrot"`
	Cheese struct {
		Cheddar bool
	}
	Sauces  SubTest
	Version string
}

type apple struct {
	text string
	Core string
}

func (apple) GET(r *Request) (interface{}, int) {
	text := "Hello"

	param := r.Vars["apple"]
	return fmt.Sprintf("%v, %v", text, param), http.StatusOK
}

func (apple) DELETE(r *Request) (interface{}, int) {
	param := r.Vars["apple"]

	text := "Deleted"

	return fmt.Sprintf("%v, %v", text, param), http.StatusAccepted
}

type SubTest struct {
	Ketchup bool
	Mayo    bool
}

func Respond(d *Data) interface{} {
	type Response struct {
		HTTPCode int
		Source   string
		Time     time.Time
		Data     interface{}
	}

	return Response{HTTPCode: d.Code, Source: d.Request.RemoteAddr, Time: time.Now(), Data: d.Data}
}

func TestGenMux(t *testing.T) {
	var d StructTest

	d.Version = "0.0.1"

	//r := GenMux("/v1/", d)
	r := GenMuxWithResponder("/v1/", d, Respond)

	log.Fatal(http.ListenAndServe(":8081", r))

}
