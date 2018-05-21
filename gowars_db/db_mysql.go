package gowars_db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"sync"
	"log"
)

var createAllTablesStatements = [][]string{
	createItemTableStatements,
	createPriceTableStatements,
}

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

type MySQLConn struct {
	Conn             *sql.DB
	MaxAllowedPacket *sql.Stmt

	Batch *Batch

	Item  *itemDatabase
	Price *priceDatabase
}

// Enforce interface
var _ IdbMysql = &MySQLConn{}

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

// newMySQLDB creates a new MySQLConn backed by a given MySQL server.
func (mysqlConfig MySQLConfig) NewMySQLDB() (*MySQLConn, error) {
	// Check database and table exists. If not, create it.
	if err := mysqlConfig.ensureTableExists(); err != nil {
		return nil, err
	}

	conn, err := sql.Open("mysql", mysqlConfig.dataStoreName("gowars"))
	//conn.SetMaxIdleConns(10)
	//conn.SetMaxOpenConns(50)
	if err != nil {
		//	return nil, fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		//	return nil, fmt.Errorf("mysql: could not establish a good connection: %v", err)
	}

	db := &MySQLConn{
		Conn: conn,
		Item: &itemDatabase{
		},
		Price: &priceDatabase{
		},
		Batch: &Batch{
			PreparedStatements: make(map[string]*sql.Stmt),
			prepstmts:          make(map[string]*sql.Stmt),
			flushInterval:      100,
			batchInserts:       make(map[string]*insert),
		},
	}

	if db.MaxAllowedPacket, err = db.Conn.Prepare(getMaxAllowedPacketStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare insert: %v", err)
	}
	db.PrepareItemStatements()
	db.PreparePriceStatements()

	return db, nil
}

// rowScanner is implemented by sql.Row and sql.Rows
type rowScanner interface {
	Scan(dest ...interface{}) error
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

	if _, err := conn.Exec("USE Item"); err != nil {
		// MySQL error 1049 is "database does not exist"
		if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == 1049 {
			return createTable(conn)
		}
	}

	if _, err := conn.Exec("DESCRIBE Item"); err != nil {
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
	for _, createStmts := range createAllTablesStatements {
		for _, createStmt := range createStmts {
			_, err := conn.Exec(createStmt)
			if err != nil {
				return err
			}
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

const getMaxAllowedPacketStatement = `SHOW VARIABLES LIKE 'max_allowed_packet';`

// Returns Database max_allowed_packet variable in bytes
func (db *MySQLConn) GetMaxAllowedPacket() (int, error) {
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

// Close closes the database, freeing up any resources.
// Before closing it finishing flushing any existing batch inserts.
func (db *MySQLConn) Close() (error) {
	var wg sync.WaitGroup

	if err := db.FlushAll(); err != nil {
		return err
	}

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		for _, stmt := range db.Batch.PreparedStatements {
			stmtCloseErr := stmt.Close()
			if stmtCloseErr != nil {
				log.Fatalln(stmtCloseErr)
			}
		}
	}(&wg)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		for _, stmt := range db.Batch.prepstmts {
			_ = stmt.Close()
		}
	}(&wg)

	wg.Wait()
	return db.Conn.Close()
}
