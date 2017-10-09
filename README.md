# mREST #
mRest is a tool for generating a [gorilla/mux](http://www.gorillatoolkit.org/pkg/mux) HTTP request multiplex, based on a specifically formatted struct.

## Quick Start ##
Initialise the package:

    package main

    import (
        "fmt"
        "log"
        "net/http"

        "github.com/supermatt/mrest/"
    )

Create your first struct. Exported fields will become lower case paths in the mux. By default, only GET requests can be performed against a field, and it's value will be returned. In this case, you can GET /version to get its value

    type api struct {
        Version string
        Hello   hello `param:"name"`
    }

If you add a param tag to your field, you can add a variable after the path. In the example above, a parameter called name is created, along with its own GET function, below, for manipulating the parameter. Running GET /hello/Foo will set the r.Vars["name"] variable to "Foo". 

    type hello string

    func (h hello) GET(r *mrest.Request) (interface{}, int) {
        greeting := string(h)
        if greeting == "" {
            greeting = "Hello, %v"
        }

        return fmt.Sprintf(greeting, r.Vars["name"]), http.StatusOK
    }

All allowed methods:
 * GET
 * POST
 * PUT
 * DELETE
 * PATCH

To generate the mux, initialise the struct and set any default values you may require. Pass this as a variable to GenMux, which also takes the root path for the urls. Finally, use that mux with net/http.

    func main() {
        var s api
        s.Version = "0.0.1"
        s.Hello = "Greetings, %v"
        mux := mrest.GenMux("/", s)

        log.Fatal(http.ListenAndServe(":8080", mux))
    }

