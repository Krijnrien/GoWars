package main

import (
	"net/http"
	"html/template"
	"github.com/krijnrien/GoWars/wrapper"
	"fmt"
)

func main() {
	tmpl := template.Must(template.ParseFiles("tmpl/amp/bank.html"))

	http.HandleFunc("/bank", func(w http.ResponseWriter, r *http.Request) {
		api, authApiErr := wrapper.NewAuthenticatedGW2Api("69D29983-607F-E143-BCD0-9DD0012AABB586611455-9747-490D-992A-CC2F188ACCF1")

		if authApiErr != nil {
			//TODO log error
			fmt.Println(authApiErr)
		}
		items, bankFetchErr := api.AccountBank()

		if bankFetchErr != nil {
			//TODO log error
			fmt.Println(authApiErr)
		}

		tmpl.Execute(w, items)
	})

	http.ListenAndServe(":8881", nil)
}
