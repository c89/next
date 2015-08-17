package next

import (
	"log"
	"testing"
)

/*
func TestMysqlParse(t *testing.T) {
	mysql := NewMysql()
	param := map[string]interface{}{
		"id":   1,
		"name": "fred",
	}
	q, err := mysql.QueryRow("SELECT * FROM buz_t WHERE id=:id AND name=:name OR uid<>:id", &param)

	if err != nil {
		t.Log(err)
		t.Error("sql parase fail")
	}

	t.Log("sql: ", q)
}
*/

var mysql *Mysql

func init() {
	mysql := NewMysql()
	mysql.Open("xcup:VX52XJbBvz7LMRQz@/xcup")
}

func TestMysqlQuery(t *testing.T) {
	out, err := mysql.Query("SELECT * FROM buz_cup")

	if err != nil {
		t.Log(err)
		t.Error("sql parase fail")
	}

	t.Log(out)
}

func TestMysqlRow(t *testing.T) {
	param := map[string]interface{}{
		"id": 1,
	}

	out, err := mysql.Row("SELECT * FROM buz_cup WHERE id=:id", &param)

	if err != nil {
		t.Log(err)
		t.Error("sql parase fail")
	}

	t.Log(out)
}

func TestMysqlExecInsert(t *testing.T) {
	maps := map[string]interface{}{
		"serial":  "123456",
		"did":     1,
		"uid":     4,
		"status":  1,
		"created": "2015-08-15",
		"updated": "2015-08-15",
	}
	out, err := mysql.Exec("buz_cup", "add", &maps)

	if err != nil {
		t.Log(err)
		t.Error("sql parase fail")
	}

	log.Print(out)
	t.Log(out)
}
func TestMysqlExecUpdate(t *testing.T) {
	u := map[string]interface{}{
		"serial":    "s123456",
		"updated:=": "now()",
	}

	w := map[string]interface{}{
		"id": "1",
	}
	out, err := mysql.Exec("buz_cup", "up", &u, &w)

	if err != nil {
		t.Log(err)
		t.Error("sql parase fail")
	}

	log.Print(out)
	t.Log(out)
}
func TestMysqlExecDelete(t *testing.T) {
	return
	w := map[string]interface{}{
		"id:=": "7",
	}
	out, err := mysql.Exec("buz_cup", "del", &w)

	if err != nil {
		t.Log(err)
		t.Error("sql parase fail")
	}

	log.Print(out)
	t.Log(out)
}
