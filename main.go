package main

import (
	"github.com/jasonlvhit/gocron"
	"log"
	"sync"
	"fmt"
	"time"
	"github.com/krijnrien/GoWars/gw2api"
	"flag"
)

func main() {
	//go cron()
	//items()
	//Api()

	var (
		maxWorkers   = flag.Int("max_workers", 1, "The number of workers to start")
		maxQueueSize = flag.Int("max_queue_size", 100, "The size of job queue")
	)
	flag.Parse()

	getPrices(maxWorkers, maxQueueSize)
}

//
func getAllPrices() {

	//TODO Start repeating 6000 ms timer
	//TODO If API request counter exceeds 50 > spawn new slave
	//TODO If timer hits 6000ms, reset API request counter

	//TODO Fetch all sellable IDS

	api := gw2api.NewGW2Api()
	allCommercePriceIds, fetchPricesErr := api.CommercePrices()
	if fetchPricesErr != nil {
		log.Fatalln(fetchPricesErr)
	}

	pageSize := 200
	totalPages := len(allCommercePriceIds) / pageSize

	var wg sync.WaitGroup
	var m sync.Mutex

	for i := 0; i <= totalPages; i++ {
		articlePrices, CommercePricePagesErr := api.CommercePricePages(i, pageSize)
		if CommercePricePagesErr != nil {
			log.Fatalln(CommercePricePagesErr)
		}

		go func(articlePrices []gw2api.ArticlePrice) {
			for _, articlePrice := range articlePrices {
				go func(articlePrice gw2api.ArticlePrice) {
					wg.Add(1)
					returnId, insertErr := insertNewCommercePrice(articlePrice, &m)
					if insertErr != nil {
						//TODO replace with log?
						fmt.Println(insertErr)
					}
					fmt.Printf("%v: inserted new item price of id: %v\n", time.Now(), returnId)
					wg.Done()
				}(articlePrice)
			}

		}(articlePrices)
	}
	wg.Wait()
	fmt.Println("Finished fetching & inserting new prices of all commerce items")
	// Fluh any remainder insert statement prices that did not hit the interval flush
	DB.FlushAll()

	//TODO Compare fetched list to memory list
	//TODO If lists size are not equal
	//TODO Add all missing items as new item details API call to next iteration

	//TODO Calculate amount of pages with page_size 200

	//TODO For loop prices pagination.
	//TODO Add API request counter

	//TODO insert price into prices table
	//TODO if item does not exist
	//TODO Keep item price in memory for next iteration
	//TODO Add missing item details API call to next iteration

}

// Setting up all cron jobs
func cron() {
	//Cron jobs do not initially call the function, only when it reached its interval.

	// Making sure that when app is started it fetches the commerce price immediatly and then on a cron job.
	//getAllCurrentCommercePrices()

	//buildCompleteDatabase()

	cron := gocron.NewScheduler()
	cron.Every(3).Hours().Do(getAllCurrentCommercePrices)
	<-cron.Start()
}

//
func buildCompleteDatabase() {
	// Build price history of all items. Takes a very long time, maybe hours.
	buildCompleteCommerceHistory()
}
