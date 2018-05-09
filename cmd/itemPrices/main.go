package main

import (
"fmt"
"github.com/krijnrien/GoWars/wrapper"
"strconv"
)

func main() {
	api := wrapper.NewGW2Api()
	b, _ := api.Items()

	for _, element := range b{
		fmt.Println(strconv.Itoa(element))
	}

	//api, _ := wrapper.NewAuthenticatedGW2Api("69D29983-607F-E143-BCD0-9DD0012AABB586611455-9747-490D-992A-CC2F188ACCF1")
	//b, _ := api.AccountBank()
	//fmt.Println(b[20].ID)

	//item := &wrapper.Item{
	//	ID:          12,
	//	Name:        "name",
	//	Description: "description",
	//	Type:        "type",
	//	Level:       68,
	//	Rarity:      "rarity",
	//	VendorValue: 234234,
	//	Icon:        "http://guildwars2.com",
	//}
	//
	//itemid, err := DB.AddItem(item)
	//
	//if err != nil {
	//	log.Fatal(err)
	//}
}
