//Package gw2api provides bindings for the Guild Wars 2 v2 API
//
//Using this package is as simple as calling the functions on the API struct
//Further examples can be found with the function definitions
//
//   func main() {
//     api := NewGW2Api()
//     b, _ := api.Build()
//     fmt.Println(b)
//   }
package gw2spidy

import (
	"net/http"
	"time"
)

// GW2Api is the state holder of the API. It includes authentication information
type GW2Spidy struct {
	timeout   time.Duration
	client    http.Client
}

// NewGW2Api returns a simple GW2Api with a HTTP timeout of 15 seconds
func NewGW2Spidy() *GW2Spidy {
	api := &GW2Spidy{
		client: http.Client{},
	}
	api.SetTimeout(time.Duration(15 * time.Second))
	return api
}

// SetTimeout set HTTP timeout for all new HTTP connections started from this
// instance
func (spidy *GW2Spidy) SetTimeout(t time.Duration) {
	spidy.timeout = t
	spidy.client.Transport = &http.Transport{
		//TODO Use DialContext
		Dial: spidy.dialTimeout,
	}
}

