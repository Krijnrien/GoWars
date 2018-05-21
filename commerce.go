package main

import (
	"log"
	"sync"
	"fmt"
	"strings"

	"github.com/krijnrien/GoWars/gw2api"
	"github.com/krijnrien/GoWars/gw2spidy"
)

type chanResult struct {
	returnId int
	error    error
}

func getAllCurrentCommercePrices() {
	api := gw2api.NewGW2Api()
	allCommercePriceIds, fetchPricesErr := api.CommercePrices()
	if fetchPricesErr != nil {
		log.Fatalln(fetchPricesErr)
	}

	pageSize := 100
	totalPages := len(allCommercePriceIds) / pageSize

	//in := make(chan []gw2api.ArticlePrice)
	//out := make(chan chanResult)
	//var m sync.Mutex


	count := 1
	for i := 0; i <= totalPages; i++ {
		articlePrices, CommercePricePagesErr := api.CommercePricePages(i, pageSize)
		if CommercePricePagesErr != nil {
			log.Fatalln(CommercePricePagesErr)
		}

		go TempinsertNewCommerPrice(articlePrices)


		//go insertNewCommerPrice(in, out, &m)
		//in <- articlePrices
		//
		//result := <-out
		//if result.error != nil {
		//	fmt.Println(result.error)
		//}

		//fmt.Printf("Fetched & inserted new commerce price of item id: %v\n", result.returnId)
		//count = count + 50
	}
	//fmt.Println(articlePrices.)
	fmt.Printf("Fetched & intersted total of %v items\n", count)
	DB.FlushAll()
}

func TempinsertNewCommerPrice(articlePrices []gw2api.ArticlePrice) {
	for _, articlePrice := range articlePrices {
		fmt.Println(articlePrice)
		//insertErr := DB.BatchInsert(gowars_db.InsertPriceNowStatement, articlePrice.ID, articlePrice.Buys.Quantity, articlePrice.Buys.UnitPrice, articlePrice.Sells.Quantity, articlePrice.Sells.Quantity)
		//if insertErr != nil {
		//	out <- chanResult{
		//		returnId: articlePrice.ID,
		//		error:    fmt.Errorf("error batchInsert: %v", insertErr),
		//	}
		//	return
		//} else{
	}
}

func insertNewCommerPrice(in <-chan []gw2api.ArticlePrice, out chan<- chanResult, m *sync.Mutex) {
	var articlePrices = <-in
	for _, articlePrice := range articlePrices {
		m.Lock()
		//insertErr := DB.BatchInsert(gowars_db.InsertPriceNowStatement, articlePrice.ID, articlePrice.Buys.Quantity, articlePrice.Buys.UnitPrice, articlePrice.Sells.Quantity, articlePrice.Sells.Quantity)
		//if insertErr != nil {
		//	out <- chanResult{
		//		returnId: articlePrice.ID,
		//		error:    fmt.Errorf("error batchInsert: %v", insertErr),
		//	}
		//	return
		//} else{
			out <- chanResult{
				returnId: articlePrice.ID,
				error:    nil,
			//}
		}
		m.Unlock()
	}
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
					error:    fmt.Errorf("error fullHistoryBuyListing: %v", fetchErr),
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
					error:    fmt.Errorf("error fullHistorySellListing: %v", fetchErr),
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
							error:    fmt.Errorf("error batchInsert: %v", insertErr),
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
