package gowars_db

import (
	"github.com/krijnrien/GoWars/gw2api"
)

type IPriceDatabase interface {
	// ListPrices returns a list of Prices
	ListPrices() ([]*gw2api.ArticlePriceTimed, error)

	//TODO description
	GetDistinctPriceHistoryIds() ([]int, error)

	// AddPrice saves a given Price, ID already assigned.
	AddPrice(b *gw2api.ArticlePriceTimed) (id int64, err error)

	//TODO description
	DropPriceTable() (err error)
}
