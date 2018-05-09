package GoWars

import (
	"github.com/krijnrien/GoWars/wrapper"
)

type ItemDatabase interface {
	// ListItems returns a list of items
	ListItems() ([]*wrapper.Item, error)

	// GetItem retrieves a book by its ID.
	GetItem(id int) (*wrapper.Item, error)

	// AddItem saves a given item, ID already assigned.
	AddItem(b *wrapper.Item) (id int64, err error)

	// DeleteItem removes a given item by its ID.
	DeleteItem(id int64) error

	// UpdateItem updates the entry for a given book.
	UpdateItem(b *wrapper.Item) error

	// Close closes the database, freeing up any available resources.
	// TODO(cbro): Close() should return an error.
	Close()
}
