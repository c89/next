package next

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"runtime"
)

type Mysql struct {
}

func NewMysql() *Mysql {
	return &Mysql{}
}

func (db *Mysql) Open(conn string) {
	conn, err = sql.Open("mysql", conn)
	if err != nil {
		panic(err.Error())
	}

	runtime.SetFinalizer(conn, db.Close)
}

func (db *Mysql) Close() {
	db.conn.Close()
}
