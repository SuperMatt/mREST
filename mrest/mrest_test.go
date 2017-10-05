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
	Sauces SubTest
}

type apple struct {
	text string
	Core string
}

func (a apple) Handler(r *http.Request, v *map[string]string) interface{} {
	vars := *v
	param := vars["apple"]
	return fmt.Sprintf("%v, %v", a.text, param)
}

type SubTest struct {
	Ketchup bool
	Mayo    bool
}

func TestGenMux(t *testing.T) {
	var d StructTest

	d.Carrot = "I am a carrot"
	d.Apple.text = "Hello"
	d.Apple.Core = "You look amazing!"

	r := GenMux("/v1/", d)

	log.Fatal(http.ListenAndServe(":8081", r))

}
