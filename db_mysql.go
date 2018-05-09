package GoWars

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/krijnrien/GoWars/wrapper"
)

var createTableStatements = []string{
	`CREATE DATABASE IF NOT EXISTS gw2 DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`,
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

// mysqlDB persists Item to a MySQL instance.
type mysqlDB struct {
	conn *sql.DB

	list        *sql.Stmt
	listBy      *sql.Stmt
	insert      *sql.Stmt
	get         *sql.Stmt
	update      *sql.Stmt
	delete      *sql.Stmt
}

// Ensure mysqlDB conforms to the ItemDatabase interface.
var _ ItemDatabase = &mysqlDB{}

type MySQLConfig struct {
	// Optional.
	Username, Password string

	// Host of the MySQL instance.
	//
	// If set, UnixSocket should be unset.
	Host string

	// Port of the MySQL instance.
	//
	// If set, UnixSocket should be unset.
	Port int

	// UnixSocket is the file path to a unix socket.
	//
	// If set, Host and Port should be unset.
	UnixSocket string
}

// dataStoreName returns a connection string suitable for sql.Open.
func (mysqlConfig MySQLConfig) dataStoreName(databaseName string) string {
	var cred string
	// [username[:password]@]
	if mysqlConfig.Username != "" {
		cred = mysqlConfig.Username
		if mysqlConfig.Password != "" {
			cred = cred + ":" + mysqlConfig.Password
		}
		cred = cred + "@"
	}

	if mysqlConfig.UnixSocket != "" {
		return fmt.Sprintf("%sunix(%s)/%s", cred, mysqlConfig.UnixSocket, databaseName)
	}
	return fmt.Sprintf("%stcp([%s]:%d)/%s", cred, mysqlConfig.Host, mysqlConfig.Port, databaseName)
}

// newMySQLDB creates a new ItemDatabase backed by a given MySQL server.
func newMySQLDB(config MySQLConfig) (ItemDatabase, error) {
	// Check database and table exists. If not, create it.
	if err := config.ensureTableExists(); err != nil {
		return nil, err
	}

	conn, err := sql.Open("mysql", config.dataStoreName("gowars"))
	if err != nil {
		return nil, fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("mysql: could not establish a good connection: %v", err)
	}

	db := &mysqlDB{
		conn: conn,
	}

	// Prepared statements. The actual SQL queries are in the code near the
	if db.list, err = conn.Prepare(listStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare list: %v", err)
	}
	if db.get, err = conn.Prepare(getStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare get: %v", err)
	}
	if db.insert, err = conn.Prepare(insertStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare insert: %v", err)
	}
	if db.update, err = conn.Prepare(updateStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare update: %v", err)
	}
	if db.delete, err = conn.Prepare(deleteStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare delete: %v", err)
	}

	return db, nil
}

// Close closes the database, freeing up any resources.
func (db *mysqlDB) Close() {
	db.conn.Close()
}

// rowScanner is implemented by sql.Row and sql.Rows
type rowScanner interface {
	Scan(dest ...interface{}) error
}

// scanItem reads a Item from a sql.Row or sql.Rows
func scanItem(s rowScanner) (*wrapper.Item, error) {
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
	if err := s.Scan(&id, &name, &description, &itemType, &level,
		&rarity, &vendorValue, &icon); err != nil {
		return nil, err
	}

	Item := &wrapper.Item{
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

const listStatement = `SELECT * FROM item`

// ListItems returns a list of Items, ordered by title.
func (db *mysqlDB) ListItems() ([]*wrapper.Item, error) {
	rows, err := db.list.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Items []*wrapper.Item
	for rows.Next() {
		Item, err := scanItem(rows)
		if err != nil {
			return nil, fmt.Errorf("mysql: could not read row: %v", err)
		}

		Items = append(Items, Item)
	}

	return Items, nil
}

const getStatement = "SELECT * FROM item WHERE id = ?"

// GetItem retrieves a Item by its ID.
func (db *mysqlDB) GetItem(id int) (*wrapper.Item, error) {
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
const insertStatement = `INSERT IGNORE INTO item (id, name, description, itemType, level, rarity, vendorValue, icon) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

// AddItem saves a given Item, assigning it a new ID.
func (db *mysqlDB) AddItem(b *wrapper.Item) (id int64, err error) {
	r, err := execAffectingOneRow(db.insert, b.ID, b.Name, b.Description, b.Type,
		b.Level, b.Rarity, b.VendorValue, b.Icon)
	if err != nil {
		return 0, err
	}

	lastInsertID, err := r.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("mysql: could not get last insert ID: %v", err)
	}
	return lastInsertID, nil
}

const deleteStatement = `DELETE FROM item WHERE id = ?`

// DeleteItem removes a given Item by its ID.
func (db *mysqlDB) DeleteItem(id int64) error {
	if id == 0 {
		return errors.New("mysql: Item with unassigned ID passed into deleteItem")
	}
	_, err := execAffectingOneRow(db.delete, id)
	return err
}

const updateStatement = `UPDATE item SET id=?, name=?, description=?, itemType=?, level=?, rarity=?, vendorValue=?, icon=? WHERE id = ?`

// UpdateItem updates the entry for a given Item.
func (db *mysqlDB) UpdateItem(b *wrapper.Item) error {
	if b.ID == 0 {
		return errors.New("mysql: Item with unassigned ID passed into updateItem")
	}

	_, err := execAffectingOneRow(db.update, b.ID, b.Name, b.Description,
		b.Type, b.Rarity, b.VendorValue, b.Icon)
	return err
}

// ensureTableExists checks the table exists. If not, it creates it.
func (mysqlConfig MySQLConfig) ensureTableExists() error {
	conn, err := sql.Open("mysql", mysqlConfig.dataStoreName("gowars"))
	if err != nil {
		return fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	defer conn.Close()

	// Check the connection.
	if conn.Ping() == driver.ErrBadConn {
		return fmt.Errorf("mysql: could not connect to the database. " +
			"could be bad address, or this address is not whitelisted for access")
	}

	if _, err := conn.Exec("USE item"); err != nil {
		// MySQL error 1049 is "database does not exist"
		if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == 1049 {
			return createTable(conn)
		}
	}

	if _, err := conn.Exec("DESCRIBE item"); err != nil {
		// MySQL error 1146 is "table does not exist"
		if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == 1146 {
			return createTable(conn)
		}
		// Unknown error.
		return fmt.Errorf("mysql: could not connect to the database: %v", err)
	}
	return nil
}

// createTable creates the table, and if necessary, the database.
func createTable(conn *sql.DB) error {
	for _, stmt := range createTableStatements {
		_, err := conn.Exec(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

// execAffectingOneRow executes a given statement, expecting one row to be affected.
func execAffectingOneRow(stmt *sql.Stmt, args ...interface{}) (sql.Result, error) {
	r, err := stmt.Exec(args...)
	if err != nil {
		return r, fmt.Errorf("mysql: could not execute statement: %v", err)
	}
	rowsAffected, err := r.RowsAffected()
	if err != nil {
		return r, fmt.Errorf("mysql: could not get rows affected: %v", err)
	} else if rowsAffected != 1 {
		return r, fmt.Errorf("mysql: expected 1 row affected, got %d", rowsAffected)
	}
	return r, nil
}
