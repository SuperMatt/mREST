package mrest

import (
	"fmt"
	"log"
	"net/http"
	"testing"
)

type StructTest struct {
	Apple  apple `param:"apple"`
	Banana int
	Carrot string
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

func (a apple) GET(r *http.Request, v *map[string]string) interface{} {
	text := "Hello"

	vars := *v
	param := vars["apple"]
	return fmt.Sprintf("%v, %v", text, param)
}

func (a apple) DELETE(r *http.Request, v *map[string]string) interface{} {
	vars := *v
	param := vars["apple"]

	text := "Deleted"

	return fmt.Sprintf("%v, %v", text, param)
}

type SubTest struct {
	Ketchup bool
	Mayo    bool
}

func TestGenMux(t *testing.T) {
	var d StructTest

	d.Version = "0.0.1"
	d.Carrot = "I am a carrot"

	r := GenMux("/v1/", d)

	log.Fatal(http.ListenAndServe(":8081", r))

}
