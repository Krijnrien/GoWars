package main

import "github.com/jasonlvhit/gocron"

func main() {
	//buildCompleteCommerceHistory()

	getAllCurrentCommercePrices()
	//
	//go cron()
	//Api()
}

func cron() {
	cron := gocron.NewScheduler()
	cron.Every(5).Minutes().Do(getAllCurrentCommercePrices)
	//cron.Every(1).Days().Do()
	<-cron.Start()
}
