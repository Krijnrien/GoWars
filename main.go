package main

import (
	"fmt"
	GoWars "github.com/krijnrien/GoWars/gw2api"
)

func main() {

	api := GoWars.NewGW2Api()
	b, _ := api.Build()
	fmt.Println(b)
}
