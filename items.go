package main

import (
	"fmt"
	"log"
	"github.com/krijnrien/GoWars/gw2api"
	"github.com/krijnrien/GoWars/gowars_db"
	"sync"
)

func items() {
	api := gw2api.NewGW2Api()
	itemIds, _ := api.Items()
	pageSize := 50
	amountOfPages := len(itemIds) / pageSize

	var wg sync.WaitGroup
	var m sync.Mutex

	for i := 0; i < amountOfPages; i++ {
		items, err := api.ItemDetails(i, pageSize, "en")
		if err != nil {
			log.Fatal(err)
		}
		go func(items []gw2api.Item) {
			for _, item := range items {
				go func(item gw2api.Item) {
					wg.Add(1)
					m.Lock()
					insertErr := DB.BatchInsert(gowars_db.InsertItemStatement, item.ID, item.Name, item.Description, item.Type, item.Level, item.Rarity, item.VendorValue, item.Icon)
					m.Unlock()
					if insertErr != nil {
						fmt.Println(insertErr)
					}
					fmt.Printf("Inserted item %v", item.ID)
					wg.Done()
				}(item)

			}
		}(items)
	}
	fmt.Println("Finished updating item database")
}
