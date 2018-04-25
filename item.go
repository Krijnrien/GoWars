package main

import (
	"github.com/krijnrien/GoWars/gw2api"
)

type ItemDatabase interface {
	// ListItems returns a list of items, alphabetically ordered by title.
	ListItems() ([]*gw2api.Item, error)

	// GetItem retrieves a book by its ID.
	GetItem(id int64) (*gw2api.Item, error)

	// AddItem saves a given item, ID already assigned.
	AddItem(b *gw2api.Item) (id int64, err error)

	// DeleteItem removes a given item by its ID.
	DeleteItem(id int64) error

	// UpdateItem updates the entry for a given book.
	UpdateItem(b *gw2api.Item) error

	// Close closes the database, freeing up any available resources.
	// TODO(cbro): Close() should return an error.
	Close()
}
