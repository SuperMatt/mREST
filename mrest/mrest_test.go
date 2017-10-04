package mrest

import (
	"fmt"
	"log"
	"net/http"
	"testing"
)

type StructTest struct {
	Apple  apple
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

func (a apple) Handler(r *http.Request) interface{} {
	return fmt.Sprintf("%v, %v", a.text, r.RemoteAddr)
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

	GenMux("/v1/", d)

	log.Fatal(http.ListenAndServe(":8081", nil))

}
