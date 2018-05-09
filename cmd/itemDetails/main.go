package main

import (
	"github.com/krijnrien/GoWars/wrapper"
	"fmt"
	"os"
	"github.com/krijnrien/GoWars"
	"log"
	"strconv"
	"html/template"
)

func main() {
	api := wrapper.NewGW2Api()
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

func loopItems(items *[]wrapper.Item) {
	for _, item := range *items {
		go addItem(&item)
	}
}

func addItem(item *wrapper.Item) {
	//fetchedItem, fetchErr := GoWars.DB.GetItem(item.ID)

	//if fetchErr != nil {
	//	log.Println(fetchErr)
	//}

	//	if fetchedItem == nil {
	_, addError := GoWars.DB.AddItem(item)
	if addError != nil {
		log.Println(addError)
	}
	fmt.Println(item.Name)
	//	parse(&item)
	//} else{
	//	fmt.Println("Item already exists in DB")
	//}
}

func parse(item *wrapper.Item) {
	path := "web/amp/" + strconv.Itoa(item.ID) + ".html"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File doesn't exist
		tmpl, err := template.ParseFiles("tmpl/amp/item.html")
		if err != nil {
			log.Print(err)
			return
		}

		file, err := os.Create(path)
		if err != nil {
			log.Println("create file: ", err)
			return
		}

		err = tmpl.Execute(file, item)
		if err != nil {
			log.Print("execute: ", err)
			return
		}
		file.Close()
	} else {
		// File exists
		fmt.Println("AMP item page already exists")
	}
}
