package gowars_db

import (
	"github.com/krijnrien/GoWars/gw2api"
)

type IItemDatabase interface {
	// ListItems returns a list of items
	ListItems() ([]*gw2api.Item, error)

	//ListItemsByName returns a list of items matching name
	ListItemsByName(name string) ([]*gw2api.Item, error)

	// GetItem retrieves a book by its ID.
	GetItem(id int) (*gw2api.Item, error)

	// AddItem saves a given Item, ID already assigned.
	AddItem(b *gw2api.Item) (id int64, err error)

	// DeleteItem removes a given Item by its ID.
	DeleteItem(id int64) error

	// UpdateItem updates the entry for a given book.
	UpdateItem(b *gw2api.Item) error

}
