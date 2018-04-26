package main

import (
	"github.com/krijnrien/GoWars/gw2api"
	"log"
	"fmt"
)

func main() {
	api := gw2api.NewGW2Api()
	itemIds, _ := api.Items()

	item2 := &gw2api.Item{
		ID:          12,
		Name:        "name",
		Description: "description",
		Type:        "type",
		Level:       68,
		Rarity:      "rarity",
		VendorValue: 234234,
		Icon:        "http://guildwars2.com",
	}

	for _, itemId := range itemIds {
		items, _ := api.ItemDetails(0, 0, "en", itemId)
		for _, item := range items {
			//fmt.Printf("%#v\n", item.Name)
			//itemid, err := DB.AddItem(item)
			

			_, err := DB.AddItem(item2)
			fmt.Println(item.Description)

			if err != nil {
				log.Fatal(err)
			}


		}
	}
}
