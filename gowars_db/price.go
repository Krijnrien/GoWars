package gowars_db

import (
	"github.com/krijnrien/GoWars/gw2api"
	"fmt"
	"database/sql"
)

var createPriceTableStatements = []string{
	`CREATE DATABASE IF NOT EXISTS gowars DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`,
	`USE gowars;`,
	`CREATE TABLE IF NOT EXISTS price (
		itemid INT NOT NULL,
  		fetched_datetime DATETIME NOT NULL,
  		buys_quantity INT NULL,
  		buys_unit_price INT NULL,
		sells_quantity INT NULL,
  		sells_unit_price INT NULL,
		PRIMARY KEY (itemid, fetched_datetime)
	)`,
}

// itemDatabase persists Item to a MySQL instance.
type priceDatabase struct {
	list                   *sql.Stmt
	distinctPiceHistoryIds *sql.Stmt
	insert                 *sql.Stmt
	dropTable              *sql.Stmt
}

func (db *MySQLConn) PreparePriceStatements() (error) {
	var err error

	// Prepared statements. The actual SQL queries are in the code near the
	if db.Price.list, err = db.Conn.Prepare(listPriceStatement); err != nil {
		return fmt.Errorf("mysql: prepare list: %v", err)
	}
	if db.Price.distinctPiceHistoryIds, err = db.Conn.Prepare(getDistinctPriceHistoryIdsStatement); err != nil {
		return fmt.Errorf("mysql: prepare get distinct price history ids: %v", err)
	}
	if db.Price.insert, err = db.Conn.Prepare(InsertPriceStatement); err != nil {
		return fmt.Errorf("mysql: prepare insert: %v", err)
	}
	if db.Price.insert, err = db.Conn.Prepare(InsertPriceNowStatement); err != nil {
		return fmt.Errorf("mysql: prepare insert: %v", err)
	}
	if db.MaxAllowedPacket, err = db.Conn.Prepare(getMaxId); err != nil {
		return fmt.Errorf("mysql: get max id: %v", err)
	}
	if db.Price.dropTable, err = db.Conn.Prepare(dropPriceTableStatement); err != nil {
		return fmt.Errorf("mysql: prepare drop: %v", err)
	}



	return nil
}

// Ensure itemDatabase conforms to the IItemDatabase interface.
var _ IPriceDatabase = &priceDatabase{}

const listPriceStatement = `SELECT * FROM price`

// ListItems returns a list of Items, ordered by title.
func (db *priceDatabase) ListPrices() ([]*gw2api.ArticlePriceTimed, error) {
	rows, err := db.list.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Items []*gw2api.ArticlePriceTimed
	for rows.Next() {
		Item, err := scanPrice(rows)
		if err != nil {
			return nil, fmt.Errorf("mysql: could not read row: %v", err)
		}

		Items = append(Items, Item)
	}

	return Items, nil
}

const getDistinctPriceHistoryIdsStatement = `SELECT DISTINCT itemid FROM price`

// ListItems returns a list of Items, ordered by title.
func (db *priceDatabase) GetDistinctPriceHistoryIds() ([]int, error) {
	rows, err := db.distinctPiceHistoryIds.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Items []int
	for rows.Next() {
		var (
			itemid int
		)
		if err := rows.Scan(&itemid); err != nil {
			return nil, err
		}

		Items = append(Items, itemid)
	}
	return Items, nil
}

const InsertPriceStatement = `INSERT INTO price (itemid, fetched_datetime, buys_quantity, buys_unit_price, sells_quantity, sells_unit_price) VALUES (?, ?, ?, ?, ?, ?)`
const InsertPriceNowStatement = `INSERT IGNORE INTO price (itemid, fetched_datetime, buys_quantity, buys_unit_price, sells_quantity, sells_unit_price) VALUES (?, now(), ?, ?, ?, ?)`

// AddItem saves a given Item, assigning it a new ID.
func (db *priceDatabase) AddPrice(b *gw2api.ArticlePriceTimed) (id int64, addErr error) {
	r, addErr := execAffectingOneRow(db.insert, b.ID, b.Fetch_datetime, b.Buys.Quantity, b.Buys.UnitPrice, b.Sells.Quantity, b.Sells.UnitPrice)
	if addErr != nil {
		return 0, addErr
	}

	lastInsertID, addErr := r.LastInsertId()
	if addErr != nil {
		return 0, fmt.Errorf("mysql: could not get last insert ID: %v", addErr)
	}
	return lastInsertID, nil
}

const getMaxId = `SELECT MAX(itemid) as max_itemid from price;`

// Returns Database max_allowed_packet variable in bytes
func (db *MySQLConn) GetMaxId() (int, error) {
	row := db.MaxAllowedPacket.QueryRow()

	var (
		column string
		value  int
	)

	if err := row.Scan(&column, &value); err != nil {
		return 0, err
	}

	return value, nil
}

const dropPriceTableStatement = `DROP TABLE price`

// AddItem saves a given Item, assigning it a new ID.
func (db *priceDatabase) DropPriceTable() (err error) {
	_, dropTableErr := execAffectingOneRow(db.dropTable)
	return dropTableErr
}

// scanItem reads a Item from a sql.Row or sql.Rows
func scanPrice(s rowScanner) (*gw2api.ArticlePriceTimed, error) {
	var (
		itemid           int
		fetched_datetime sql.NullString
		buys_quantity    int
		buys_unit_price  int
		sells_quantity   int
		sells_unit_price int
	)
	if err := s.Scan(&itemid, &fetched_datetime, &buys_quantity, &buys_unit_price, &sells_quantity, &sells_unit_price); err != nil {
		return nil, err
	}

	Item := &gw2api.ArticlePriceTimed{
		ID:             itemid,
		Fetch_datetime: fetched_datetime.String,
		Buys: gw2api.Price{
			Quantity:  buys_quantity,
			UnitPrice: buys_unit_price,
		},
		Sells: gw2api.Price{
			Quantity:  sells_quantity,
			UnitPrice: sells_unit_price,
		},
	}
	return Item, nil
}
