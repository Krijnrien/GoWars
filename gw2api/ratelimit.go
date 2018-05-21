package gw2api

import "time"

var semaphore = make(chan struct{}, 5)
var rate = make(chan struct{}, 10)



func init(){
	// leaky bucket
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			_, ok := <-rate
			// if this isn't going to run indefinitely, signal
			// this to return by closing the rate channel.
			if !ok {
				return
			}
		}
	}()

}