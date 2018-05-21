package main

import (
	"fmt"
	"log"
	"github.com/krijnrien/GoWars/gw2api"
)

func items() {
	api := gw2api.NewGW2Api()
	itemIds, _ := api.Items()

	//for _, itemId := range itemIds {
	//	items, _ := api.ItemDetails(0, 50, "en")
	//	go loopItems(&items)
	//}

	pageSize := 50

	amountOfPages := len(itemIds) / pageSize

	for i := 0; i < amountOfPages; i++ {
		items, err := api.ItemDetails(i, pageSize, "en")
		if err != nil {
			log.Fatal(err)
		}
		loopItems(&items)
	}
}

func loopItems(items *[]gw2api.Item) {
	for _, item := range *items {
		go addItem(&item)
	}
}

func addItem(item *gw2api.Item) {
	//fetchedItem, fetchErr := GoWars.DB.GetItem(item.ID)

	//if fetchErr != nil {
	//	log.Println(fetchErr)
	//}

	//	if fetchedItem == nil {
	_, addError := DB.Item.AddItem(item)
	if addError != nil {
		log.Println(addError)
	}
	fmt.Println(item.Name)
	//	parse(&item)
	//} else{
	//	fmt.Println("Item already exists in DB")
	//}
}
