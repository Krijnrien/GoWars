package gw2spidy

type Listings struct {
	SellOrBuy string `json:"sell-or-buy"`
	Count     int    `json:"count"`
	Page      int    `json:"page"`
	LastPage  int    `json:"last_page"`
	Total     int    `json:"total"`
	Results   []struct {
		ListingDatetime string `json:"listing_datetime"`
		UnitPrice       int    `json:"unit_price"`
		Quantity        int    `json:"quantity"`
		Listings        int    `json:"listings"`
	} `json:"results"`
}

func (spidy *GW2Spidy) ListingsSellId(id int, page int) (itemListings Listings, err error) {
	ver := "v0.9"
	datatype := "json"
	tag := "listings"
	spec := "sell"
	err = spidy.fetchEndpoint(ver, datatype, tag, id, spec, page, &itemListings)
	return
}

func (spidy *GW2Spidy) ListingsBuyId(id int, page int) (itemListings Listings, err error) {
	ver := "v0.9"
	datatype := "json"
	tag := "listings"
	spec := "buy"
	err = spidy.fetchEndpoint(ver, datatype, tag, id, spec, page, &itemListings)
	return
}
