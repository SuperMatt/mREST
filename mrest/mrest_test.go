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

func (apple) GET(r *Request) interface{} {
	text := "Hello"

	param := r.Vars["apple"]
	return fmt.Sprintf("%v, %v", text, param)
}

func (apple) DELETE(r *Request) interface{} {
	param := r.Vars["apple"]

	text := "Deleted"

	return fmt.Sprintf("%v, %v", text, param)
}

type SubTest struct {
	Ketchup bool
	Mayo    bool
}

func Respond(c int, i interface{}) interface{} {
	type Response struct {
		HTTPCode int
		Time     time.Time
		Data     interface{}
	}

	return Response{HTTPCode: c, Time: time.Now(), Data: i}
}

func TestGenMux(t *testing.T) {
	var d StructTest

	d.Version = "0.0.1"

	r := GenMux("/v1/", d, Respond)

	log.Fatal(http.ListenAndServe(":8081", r))

}
