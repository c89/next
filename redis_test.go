package next

import (
	"log"
	"testing"
)

var rds *Redis

func init() {
	rds = NewRedis()
	rds.Pool(":6379")
}

func TestDo(t *testing.T) {
	reply, err := rds.Do("SET", "say", "hello")

	if err != nil {
		t.Log(err)
		t.Error("Redis do fail")
	}
	log.Print(reply)

	reply, err = rds.Do("GET", "say")
	if err != nil {
		t.Log(err)
		t.Error("Redis do fail")
	}
	log.Printf("Get say: %T, %s", reply, reply)

	t.Log(reply)
}
