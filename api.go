package main

import (
	"github.com/julienschmidt/httprouter"

	"net/http"
	"log"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/krijnrien/GoWars/gw2api"
)

func Api() {
	port := "8082"
	router := httprouter.New()
	router.GET("/item", ListByName)
	router.GET("/items", ListAll)
	router.GET("/item/:id", ById)

	router.GET("/account/bank", AccountListBank)

	fmt.Println("http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func ById(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	id, _ := strconv.Atoi(ps.ByName("id"))
	item, fetchError := DB.Item.GetItem(id)
	if fetchError != nil {
		fmt.Println(fetchError)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(item)
}

func ListAll(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	items, fetchError := DB.Item.ListItems()
	if fetchError != nil {
		fmt.Println(fetchError)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(items)
}

func ListByName(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	queryName := r.URL.Query().Get("name")
	items, fetchError := DB.Item.ListItemsByName(queryName)
	if fetchError != nil {
		fmt.Println(fetchError)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(items)
}

func AccountListBank(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	key := r.URL.Query().Get("key")

	api, authErr := gw2api.NewAuthenticatedGW2Api(key)
	if authErr != nil {
		fmt.Println(authErr)
	}

	accountItems, _ := api.AccountBank()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(accountItems)
	//TODO Above is only boiler code, nothing more
}
