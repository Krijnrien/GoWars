package main

import (
	"fmt"
	gw2api "github.com/krijnrien/GoWars/gw2api"
)

func main() {

	api := gw2api.NewGW2Api()
	b, _ := api.Build()
	fmt.Println(b)
}
