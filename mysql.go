package next

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

type Mysql struct {
	Db *sql.DB
	Tx *sql.Tx
}

func NewMysql() *Mysql {
	return &Mysql{}
}

func (mysql *Mysql) Open(dsn string) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err.Error())
	}

	mysql.Db = db
	runtime.SetFinalizer(db, nil)
}

func (mysql *Mysql) Ping() error {
	return mysql.Ping()
}

func (mysql *Mysql) Close() {
	mysql.Db.Close()
}

func (mysql *Mysql) Row(args ...interface{}) (interface{}, error) {
	// Parse sql
	q, param, err := mysql.parse(args...)
	if err != nil {
		return nil, err
	}

	var stmt *sql.Stmt
	// Prepare
	if mysql.Tx != nil {
		stmt, err = mysql.Tx.Prepare(q)
	} else {
		stmt, err = mysql.Db.Prepare(q)
	}
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// -------------
	// Scan
	// -------------
	rows, err := stmt.Query(param...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get all col name
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Scan value
	out := make(map[string]interface{}, len(cols))
	if rows.Next() {
		rbs := make([]sql.RawBytes, len(cols))
		vals := make([]interface{}, len(cols))
		for i := range vals {
			vals[i] = &rbs[i]
		}

		if err = rows.Scan(vals...); err != nil {
			return nil, err
		}

		// Make map
		for i, val := range rbs {
			out[cols[i]] = val
		}
	}

	return out, nil
}

func (mysql *Mysql) Query(args ...interface{}) (interface{}, error) {
	// Parse sql
	q, param, err := mysql.parse(args...)
	if err != nil {
		return nil, err
	}

	var stmt *sql.Stmt
	// Prepare
	if mysql.Tx != nil {
		stmt, err = mysql.Tx.Prepare(q)
	} else {
		stmt, err = mysql.Db.Prepare(q)
	}
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// -------------
	// Scan
	// -------------
	rows, err := stmt.Query(param...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get all col name
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Scan value
	out := make([]interface{}, 0)
	for rows.Next() {
		rbs := make([]sql.RawBytes, len(cols))
		vals := make([]interface{}, len(cols))
		for i := range vals {
			vals[i] = &rbs[i]
		}

		if err = rows.Scan(vals...); err != nil {
			return nil, err
		}

		// Make map
		r := make(map[string]interface{}, len(cols))
		for i, val := range rbs {
			r[cols[i]] = val
		}
		out = append(out, r)
	}

	log.Print(out)
	return out, nil
}

func (mysql *Mysql) Exec(table, dml string, data map[string]interface{}) (interface{}, error) {
	//
	switch strings.ToLower(dml); {
	case dml == "add":
		/*
			stmt, err := db.Prepare("INSERT INTO users(name) VALUES(?)")
			if err != nil {
			log.Fatal(err)
			}
			res, err := stmt.Exec("Dolly")
			if err != nil {
			log.Fatal(err)
			}
			lastId, err := res.LastInsertId()
			if err != nil {
			log.Fatal(err)
			}
			rowCnt, err := res.RowsAffected()
			if err != nil {
			log.Fatal(err)
			}
			log.Printf("ID = %d, affected = %d\n", lastId, rowCnt)
		*/
	case dml == "up":
	case dml == "del":
	}
	return nil, nil
}

func (mysql *Mysql) Begin() {
	tx, err := mysql.Db.Begin()
	if err != nil {
		panic(err.Error())
	}
	mysql.Tx = tx
}

func (mysql *Mysql) Rollback() error {
	if mysql.Tx != nil {
		return mysql.Tx.Rollback()
	}
	return nil
}

func (mysql *Mysql) Commit() error {
	if mysql.Tx != nil {
		return mysql.Tx.Commit()
	}
	return nil
}

func (mysql *Mysql) parse(args ...interface{}) (string, []interface{}, error) {
	// SQL
	q, ok := (args[0]).(string)
	if !ok {
		return "", []interface{}{}, errors.New("first param should be a sql string")
	}

	// No param
	if len(args) == 1 {
		return q, nil, nil
	}

	// Parse sql with mapping
	v := reflect.ValueOf(args[1])
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Map {
		return "", []interface{}{}, errors.New("param should be a map's pointer")
	}

	param := make([]interface{}, 0)
	re := regexp.MustCompile(":\\w+")
	q = re.ReplaceAllStringFunc(q, func(src string) string {
		param = append(param, v.Elem().MapIndex(reflect.ValueOf(src[1:])).Interface())
		return "?"
	})

	return q, param, nil
}
