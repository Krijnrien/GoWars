package gowars_db

import (
	"database/sql"
	"regexp"
	"strings"
)

var (
	dupeRegexp   = regexp.MustCompile(`(?i)on duplicate key update`)
	valuesRegexp = regexp.MustCompile(`(?i)values`)
)

// Batch is a database handle that embeds the standard library's sql.Batch struct.
//
//This means the fastsql.Batch struct has, and allows, access to all of the standard library functionality while also providng a superset of functionality such as batch operations, autmatically created prepared statmeents, and more.
type Batch struct {
	PreparedStatements map[string]*sql.Stmt
	prepstmts          map[string]*sql.Stmt
	driverName         string
	flushInterval      uint
	batchInserts       map[string]*insert
}

// BatchInsert takes a singlular INSERT query and converts it to a batch-insert query for the caller.  A batch-insert is ran every time BatchInsert is called a multiple of flushInterval times.
func (db *MySQLConn) BatchInsert(query string, params ...interface{}) (err error) {
	if _, ok := db.Batch.batchInserts[query]; !ok {
		db.Batch.batchInserts[query] = newInsert()
	} //if

	// Only split out query the first time Insert is called
	if db.Batch.batchInserts[query].queryPart1 == "" {
		db.Batch.batchInserts[query].splitQuery(query)
	}

	db.Batch.batchInserts[query].insertCtr++

	// Build VALUES seciton of query and add to parameter slice
	db.Batch.batchInserts[query].values += db.Batch.batchInserts[query].queryPart2
	db.Batch.batchInserts[query].bindParams = append(db.Batch.batchInserts[query].bindParams, params...)

	// If the batch interval has been hit, execute a batch insert
	if db.Batch.batchInserts[query].insertCtr >= db.Batch.flushInterval {
		err = db.flushInsert(db.Batch.batchInserts[query])
	} //if

	return err
}

// FlushAll iterates over all batch inserts and inserts them into the database.
func (db *MySQLConn) FlushAll() error {
	for _, in := range db.Batch.batchInserts {
		if err := db.flushInsert(in); err != nil {
			return err
		}
	}

	return nil
}

// flushInsert performs the acutal batch-insert query.
func (db *MySQLConn) flushInsert(in *insert) error {
	var (
		err   error
		query = in.queryPart1 + in.values[:len(in.values)-1] + in.queryPart3
	)

	// Prepare query
	if _, ok := db.Batch.prepstmts[query]; !ok {
		var stmt *sql.Stmt

		if stmt, err = db.Conn.Prepare(query); err == nil {
			db.Batch.prepstmts[query] = stmt
		} else {
			return err
		}
	}

	// Executate batch insert
	if _, err = db.Batch.prepstmts[query].Exec(in.bindParams...); err != nil {
		return err
	} //if

	// Reset vars
	in.values = " VALUES"
	in.bindParams = make([]interface{}, 0)
	in.insertCtr = 0

	return err
}

type insert struct {
	bindParams []interface{}
	insertCtr  uint
	queryPart1 string
	queryPart2 string
	queryPart3 string
	values     string
}

func newInsert() *insert {
	return &insert{
		bindParams: make([]interface{}, 0),
		values:     " VALUES",
	}
}

func (in *insert) splitQuery(query string) {
	var (
		ndxOnDupe, ndxValues = -1, -1
		ndxParens            = strings.LastIndex(query, ")")
	)

	// Find "VALUES".
	valuesMatches := valuesRegexp.FindStringIndex(query)
	if len(valuesMatches) > 0 {
		ndxValues = valuesMatches[0]
	}

	// Find "ON DUPLICATE KEY UPDATE"
	dupeMatches := dupeRegexp.FindAllStringIndex(query, -1)
	if len(dupeMatches) > 0 {
		ndxOnDupe = dupeMatches[len(dupeMatches)-1][0]
	}

	// Split out first part of query
	in.queryPart1 = strings.TrimSpace(query[:ndxValues])

	// If ON DUPLICATE clause exists, separate into 3 parts.
	// If ON DUPLICATE does not exist, seperate into 2 parts.
	if ndxOnDupe != -1 {
		in.queryPart2 = query[ndxValues+6:ndxOnDupe-1] + ","
		in.queryPart3 = query[ndxOnDupe:]
	} else {
		in.queryPart2 = query[ndxValues+6:ndxParens+1] + ","
	}
}
