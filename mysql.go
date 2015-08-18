package next

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
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
	return mysql.Db.Ping()
}

func (mysql *Mysql) Close() {
	mysql.Db.Close()
}

func (mysql *Mysql) Query(args ...interface{}) ([]interface{}, error) {
	// Parse sql
	q, param, err := mysql.parseQuery(args...)
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
			r[cols[i]] = string(val)
		}
		out = append(out, r)
	}

	return out, nil
}

func (mysql *Mysql) Row(args ...interface{}) (interface{}, error) {
	// Parse sql
	q, param, err := mysql.parseQuery(args...)
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
			out[cols[i]] = string(val)
		}
	}

	return out, nil
}

// DML for tabel, insert update delete operation.
// usage:
// mysql.Exec("users", "add", &addData)
// mysql.Exec("users", "up", &upData, &where)
// mysql.Exec("users", "del", &where)
func (mysql *Mysql) Exec(table, dml string, args ...interface{}) (interface{}, error) {
	if len(table) == 0 {
		panic("plase set table for exec method")
	}

	switch strings.ToLower(dml); dml {
	case "add":
		return mysql.insert(table, args...)
	case "up":
		return mysql.update(table, args...)
	case "del":
		return mysql.del(table, args...)
	default:
		panic("mysql exec method <add, up, del> can not support: " + dml)
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

func (mysql *Mysql) parseQuery(args ...interface{}) (string, []interface{}, error) {
	// SQL
	q, ok := (args[0]).(string)
	if !ok {
		panic("first param should be a sql string")
	}

	// No param
	if len(args) == 1 {
		return q, nil, nil
	}

	// Parse sql with mapping
	v := reflect.ValueOf(args[1])
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Map {
		panic("param should be a map's pointer")
	}

	param := make([]interface{}, 0)
	re := regexp.MustCompile(":\\w+")
	q = re.ReplaceAllStringFunc(q, func(src string) string {
		param = append(param, v.Elem().MapIndex(reflect.ValueOf(src[1:])).Interface())
		return "?"
	})

	return q, param, nil
}

func (mysql *Mysql) insert(table string, args ...interface{}) (interface{}, error) {
	v := reflect.ValueOf(args[0])
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Map {
		panic("param should be a map's pointer")
	}

	// Parse args
	maps := v.Elem().MapKeys()
	feild := make([]string, len(maps))
	bind := make([]string, len(maps))
	param := make([]interface{}, len(maps))
	for i, k := range maps {
		feild[i] = fmt.Sprintf("`%s`", k)
		bind[i] = "?"
		param[i] = v.Elem().MapIndex(k).Interface()
	}
	q := fmt.Sprintf("INSERT INTO `%s`(%s) VALUES(%s);", table, strings.Join(feild, ","), strings.Join(bind, ","))

	// Sql execute
	var stmt *sql.Stmt
	var err error
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

	res, err := stmt.Exec(param...)
	if err != nil {
		return nil, err
	}

	return res.LastInsertId()
}

func (mysql *Mysql) update(table string, args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		panic("mysql update method need the data of update and where")
	}

	u := reflect.ValueOf(args[0])
	if u.Kind() != reflect.Ptr || u.Elem().Kind() != reflect.Map {
		panic("updata args should be a map's pointer")
	}
	w := reflect.ValueOf(args[1])
	if w.Kind() != reflect.Ptr || w.Elem().Kind() != reflect.Map {
		panic("where args should be a map's pointer")
	}

	param := make([]interface{}, 0)
	// Parse upate args
	mu := u.Elem().MapKeys()
	up := make([]string, len(mu))
	for i, k := range mu {
		f := strings.SplitN(k.String(), ":", 2)
		if len(f) > 1 {
			up[i] = fmt.Sprintf("`%s`%s%s", f[0], f[1], u.Elem().MapIndex(k).Interface())
		} else {
			up[i] = fmt.Sprintf("`%s`=?", k)
			param = append(param, u.Elem().MapIndex(k).Interface())
		}
	}
	// Parse upate args
	mw := w.Elem().MapKeys()
	where := make([]string, len(mw))
	for i, k := range mw {
		f := strings.SplitN(k.String(), ":", 2)
		if len(f) > 1 {
			where[i] = fmt.Sprintf("`%s`%s%s", f[0], f[1], w.Elem().MapIndex(k).Interface())
		} else {
			where[i] = fmt.Sprintf("`%s`=?", k)
			param = append(param, w.Elem().MapIndex(k).Interface())
		}
	}
	q := fmt.Sprintf("UPDATE `%s` SET %s WHERE %s;", table, strings.Join(up, ","), strings.Join(where, " AND "))

	// Sql execute
	var stmt *sql.Stmt
	var err error
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

	res, err := stmt.Exec(param...)
	if err != nil {
		return nil, err
	}

	return res.RowsAffected()
}

func (mysql *Mysql) del(table string, args ...interface{}) (interface{}, error) {
	w := reflect.ValueOf(args[0])
	if w.Kind() != reflect.Ptr || w.Elem().Kind() != reflect.Map {
		panic("where args should be a map's pointer")
	}

	mw := w.Elem().MapKeys()
	param := make([]interface{}, 0)
	// Parse where args
	where := make([]string, len(mw))
	for i, k := range mw {
		f := strings.SplitN(k.String(), ":", 2)
		if len(f) > 1 {
			where[i] = fmt.Sprintf("`%s`%s%s", f[0], f[1], w.Elem().MapIndex(k).Interface())
		} else {
			where[i] = fmt.Sprintf("`%s`=?", k)
			param = append(param, w.Elem().MapIndex(k).Interface())
		}
	}
	q := fmt.Sprintf("DELETE FROM `%s` WHERE %s;", table, strings.Join(where, " AND "))

	// Sql execute
	var stmt *sql.Stmt
	var err error
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

	res, err := stmt.Exec(param...)
	if err != nil {
		return nil, err
	}

	return res.RowsAffected()
}
