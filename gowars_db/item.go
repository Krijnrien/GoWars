package gowars_db

import (
	"github.com/krijnrien/GoWars/gw2api"
	"fmt"
	"database/sql"
	"errors"
)

var createItemTableStatements = []string{
	`CREATE DATABASE IF NOT EXISTS gowars DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`,
	`USE gowars;`,
	`CREATE TABLE IF NOT EXISTS item (
		id INT UNSIGNED NOT NULL UNIQUE,
		name VARCHAR(255) NULL,
		description VARCHAR(255) NULL,
		itemType VARCHAR(255) NULL,
		level INT NULL,
		rarity TEXT NULL,
		vendorValue INT NULL,
		icon VARCHAR(255) NULL,
		PRIMARY KEY (id)
	)`,
}

// itemDatabase persists Item to a MySQL instance.
type itemDatabase struct {
	list       *sql.Stmt
	listByName *sql.Stmt
	insert     *sql.Stmt
	get        *sql.Stmt
	update     *sql.Stmt
	delete     *sql.Stmt
}

// Ensure itemDatabase conforms to the IItemDatabase interface.
var _ IItemDatabase = &itemDatabase{}

func (db *MySQLConn) PrepareItemStatements() (error) {
	var err error

	// Prepared statements. The actual SQL queries are in the code near the
	if db.Item.list, err = db.Conn.Prepare(listItemsStatement); err != nil {
		return fmt.Errorf("mysql: prepare list: %v", err)
	}
	if db.Item.listByName, err = db.Conn.Prepare(listItemsByNameStatement); err != nil {
		return fmt.Errorf("mysql: prepare list: %v", err)
	}
	if db.Item.get, err = db.Conn.Prepare(getItemStatement); err != nil {
		return fmt.Errorf("mysql: prepare get: %v", err)
	}
	if db.Item.insert, err = db.Conn.Prepare(insertItemStatement); err != nil {
		return fmt.Errorf("mysql: prepare insert: %v", err)
	}
	if db.Item.update, err = db.Conn.Prepare(updateItemStatement); err != nil {
		return fmt.Errorf("mysql: prepare update: %v", err)
	}
	if db.Item.delete, err = db.Conn.Prepare(deleteItemStatement); err != nil {
		return fmt.Errorf("mysql: prepare delete: %v", err)
	}
	return nil
}

const listItemsStatement = `SELECT * FROM item`

// ListItems returns a list of Items, ordered by title.
func (db *itemDatabase) ListItems() ([]*gw2api.Item, error) {
	rows, err := db.list.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Items []*gw2api.Item
	for rows.Next() {
		Item, err := scanItem(rows)
		if err != nil {
			return nil, fmt.Errorf("mysql: could not read row: %v", err)
		}

		Items = append(Items, Item)
	}

	return Items, nil
}

const listItemsByNameStatement = "SELECT * FROM item WHERE name LIKE ?"

// ListItems returns a list of Items, ordered by title.
func (db *itemDatabase) ListItemsByName(name string) ([]*gw2api.Item, error) {
	q := "%" + name + "%"
	rows, err := db.listByName.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Items []*gw2api.Item
	for rows.Next() {
		Item, err := scanItem(rows)
		if err != nil {
			return nil, fmt.Errorf("mysql: could not read row: %v", err)
		}

		Items = append(Items, Item)
	}

	return Items, nil
}

const getItemStatement = "SELECT * FROM item WHERE id = ?"

// GetItem retrieves a Item by its ID.
func (db *itemDatabase) GetItem(id int) (*gw2api.Item, error) {
	Item, err := scanItem(db.get.QueryRow(id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("mysql: could not find Item with id %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("mysql: could not get Item: %v", err)
	}
	return Item, nil
}

//TODO Ignoring duplication error for now, check before inserting if exists or values changes?
const insertItemStatement = `INSERT IGNORE INTO item (id, name, description, itemType, level, rarity, vendorValue, icon) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

// AddItem saves a given Item, assigning it a new ID.
func (db *itemDatabase) AddItem(b *gw2api.Item) (id int64, err error) {
	r, err := execAffectingOneRow(db.insert, b.ID, b.Name, b.Description, b.Type, b.Level, b.Rarity, b.VendorValue, b.Icon)
	if err != nil {
		return 0, err
	}

	lastInsertID, err := r.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("mysql: could not get last insert ID: %v", err)
	}
	return lastInsertID, nil
}

const deleteItemStatement = `DELETE FROM item WHERE id = ?`

// DeleteItem removes a given Item by its ID.
func (db *itemDatabase) DeleteItem(id int64) error {
	if id == 0 {
		return errors.New("mysql: Item with unassigned ID passed into deleteItem")
	}
	_, err := execAffectingOneRow(db.delete, id)
	return err
}

const updateItemStatement = `UPDATE item SET id=?, name=?, description=?, itemType=?, level=?, rarity=?, vendorValue=?, icon=? WHERE id = ?`

// UpdateItem updates the entry for a given Item.
func (db *itemDatabase) UpdateItem(b *gw2api.Item) error {
	if b.ID == 0 {
		return errors.New("mysql: Item with unassigned ID passed into updateItem")
	}

	_, err := execAffectingOneRow(db.update, b.ID, b.Name, b.Description,
		b.Type, b.Rarity, b.VendorValue, b.Icon)
	return err
}

// scanItem reads a Item from a sql.Row or sql.Rows
func scanItem(s rowScanner) (*gw2api.Item, error) {
	var (
		id          int
		name        sql.NullString
		description sql.NullString
		itemType    sql.NullString
		level       int
		rarity      sql.NullString
		vendorValue int
		icon        sql.NullString
	)
	if err := s.Scan(&id, &name, &description, &itemType, &level, &rarity, &vendorValue, &icon); err != nil {
		return nil, err
	}

	Item := &gw2api.Item{
		ID:          id,
		Name:        name.String,
		Description: description.String,
		Type:        itemType.String,
		Level:       level,
		Rarity:      description.String,
		VendorValue: vendorValue,
		Icon:        icon.String,
	}
	return Item, nil
}
