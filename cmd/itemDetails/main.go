package main

import (
	"github.com/krijnrien/GoWars/gw2api"
	"log"
)

func main() {
	api := gw2api.NewGW2Api()
	itemIds, _ := api.Items()

	for _, itemId := range itemIds {
		items, _ := api.ItemDetails(0, 0, "en", itemId)
		for _, item := range items {
			//fmt.Printf("%#v\n", item.Name)
			//itemid, err := DB.AddItem(item)
			

			itemid, err := DB.AddItem(item)

			if err != nil {
				log.Fatal(err)
			}


		}
	}
}
