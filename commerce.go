package main

import (
	"log"
	"sync"
	"fmt"
	"strings"

	"github.com/krijnrien/GoWars/gw2api"
	"github.com/krijnrien/GoWars/gw2spidy"
	"github.com/krijnrien/GoWars/gowars_db"
	"time"
	"os"
	"encoding/csv"
	"strconv"
)

type chanResult struct {
	returnId int
	error    error
}

type job struct {
	articlePrice gw2api.ArticlePrice
	write        *csv.Writer
}

var mu sync.Mutex
var prices []gw2api.ArticlePrice

func doWork(j job) {
	mu.Lock()
	defer mu.Unlock()
	s := []string{
		strconv.Itoa(j.articlePrice.ID),
		strconv.Itoa(j.articlePrice.Sells.UnitPrice),
		strconv.Itoa(j.articlePrice.Sells.Quantity),
		strconv.Itoa(j.articlePrice.Buys.UnitPrice),
		strconv.Itoa(j.articlePrice.Buys.Quantity),
	}

	j.write.Write(s)
}

func getPrices(maxQueueSize *int, maxWorkers *int) {
	defer timeTrack(time.Now(), "main")

	// create job channel
	jobs := make(chan job, *maxQueueSize)

	// create workers
	for i := 1; i <= *maxWorkers; i++ {
		go func(i int) {
			for j := range jobs {
				doWork(j)
			}
		}(i)
	}

	api := gw2api.NewGW2Api()
	allCommercePriceIds, fetchPricesErr := api.CommercePrices()
	if fetchPricesErr != nil {
		log.Fatalln(fetchPricesErr)
	}

	pageSize := 200
	totalPages := len(allCommercePriceIds) / pageSize

	file, err := os.OpenFile("result.csv", os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for i := 0; i <= totalPages; i++ {
		articlePrices, CommercePricePagesErr := api.CommercePricePages(i, pageSize)
		if CommercePricePagesErr != nil {
			log.Fatalln(CommercePricePagesErr)
		}
		for _, articlePrice := range articlePrices {
			go func(articlePrice gw2api.ArticlePrice) {
				jobs <- job{articlePrice, writer}
			}(articlePrice)
		}
	}
	fmt.Println(len(prices))
}

func getAllCurrentCommercePrices() {
	api := gw2api.NewGW2Api()
	allCommercePriceIds, fetchPricesErr := api.CommercePrices()
	if fetchPricesErr != nil {
		log.Fatalln(fetchPricesErr)
	}

	//TODO Make config variable, ENV?
	pageSize := 100
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
}

func insertNewCommercePrice(articlePrice gw2api.ArticlePrice, m *sync.Mutex) (int, error) {
	m.Lock()
	insertErr := DB.BatchInsert(gowars_db.InsertPriceNowStatement, articlePrice.ID, articlePrice.Buys.Quantity, articlePrice.Buys.UnitPrice, articlePrice.Sells.Quantity, articlePrice.Sells.Quantity)
	m.Unlock()
	if insertErr != nil {
		return articlePrice.ID, fmt.Errorf("error batchInsert: %v\n", insertErr)
	}
	return articlePrice.ID, nil
}

// Build complete commerce price history of all items. Is only supposed to once, populating the database.
func buildCompleteCommerceHistory() {
	api := gw2api.NewGW2Api()
	allCommerceIds, fetchCommerceIdsErr := api.CommercePrices()
	if fetchCommerceIdsErr != nil {
		log.Fatalln(fetchCommerceIdsErr)
	}
	fmt.Println(len(allCommerceIds))

	distinctExistingPriceHistoryIds, fetchDistinctExistingPriceIds := DB.Price.GetDistinctPriceHistoryIds()
	if fetchDistinctExistingPriceIds != nil {
		log.Fatalln(fetchDistinctExistingPriceIds)
	}

	in := make(chan int)
	out := make(chan chanResult)
	var m sync.Mutex

	for _, id := range allCommerceIds {
		if !Contains(distinctExistingPriceHistoryIds, id) {
			go insertFullCommerceHistoryById(in, out, &m)
			in <- id

			result := <-out
			if result.error != nil {
				fmt.Println(result.error)
				//close(result)
			}
			fmt.Println(result.returnId)
		}
	}
	// Flushing last item inserts that didn't hit the 100 interval mark.
	DB.FlushAll()
}

// inserting commerce prices by bulk insert
func insertFullCommerceHistoryById(in <-chan int, out chan<- chanResult, m *sync.Mutex) {
	var wg sync.WaitGroup
	wg.Add(2)

	spidy := gw2spidy.NewGW2Spidy()
	var fullHistoryBuyListing gw2spidy.Listings
	var fullHistorySellListing gw2spidy.Listings
	var cancelerr = false
	var id = <-in
	go func() {
		var lastPage = 1

		for i := 1; i <= lastPage; i++ {
			items, fetchErr := spidy.ListingsBuyId(id, i)
			lastPage = items.LastPage

			if fetchErr != nil {
				cancelerr = true
				out <- chanResult{
					returnId: id,
					error:    fmt.Errorf("error fullHistoryBuyListing: %v\n", fetchErr),
				}
				return
			}
			fullHistoryBuyListing.Results = append(fullHistoryBuyListing.Results, items.Results...)
		}
		defer wg.Done()
	}()

	go func() {
		var lastPage = 1
		for i := 1; i <= lastPage; i++ {
			items, fetchErr := spidy.ListingsSellId(id, i)
			lastPage = items.LastPage

			if fetchErr != nil {
				cancelerr = true
				out <- chanResult{
					returnId: id,
					error:    fmt.Errorf("error fullHistorySellListing: %v\n", fetchErr),
				}
				return
			}
			fullHistorySellListing.Results = append(fullHistorySellListing.Results, items.Results...)
		}
		defer wg.Done()
	}()

	wg.Wait()
	if cancelerr == false {
		for _, buy := range fullHistoryBuyListing.Results {
			for _, sell := range fullHistorySellListing.Results {
				if buy.ListingDatetime == sell.ListingDatetime {
					ArticlePriceTimed := gw2api.ArticlePriceTimed{
						ID:             id,
						Fetch_datetime: buy.ListingDatetime,
						Buys: gw2api.Price{
							Quantity:  sell.Quantity,
							UnitPrice: sell.UnitPrice,
						},
						Sells: gw2api.Price{
							Quantity:  buy.Quantity,
							UnitPrice: buy.UnitPrice,
						},
					}

					m.Lock()
					insertErr := DB.BatchInsert("INSERT IGNORE INTO price (itemid, fetched_datetime, buys_quantity, buys_unit_price, sells_quantity, sells_unit_price) VALUES (?, ?, ?, ?, ?, ?);", ArticlePriceTimed.ID, strings.TrimSuffix(ArticlePriceTimed.Fetch_datetime, " UTC"), ArticlePriceTimed.Buys.Quantity, ArticlePriceTimed.Buys.UnitPrice, ArticlePriceTimed.Sells.Quantity, ArticlePriceTimed.Sells.Quantity)
					if insertErr != nil {
						out <- chanResult{
							returnId: id,
							error:    fmt.Errorf("error batchInsert: %v\n", insertErr),
						}
						return
					}
					m.Unlock()

				}
			}
		}
	}
	out <- chanResult{
		returnId: id,
		error:    nil,
	}
	return
}
