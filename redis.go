package next

/*
import (
	"github.com/garyburd/redigo/redis"
)

type Conn struct {
	redis.Conn
}

func (con Conn) Close() {
	con.Conn.Close()
}

func initRedis(host string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     64,
		IdleTimeout: 60 * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")

			return err
		},
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host)
			if err != nil {
				return nil, err
			}

			_, err = c.Do("SELECT", config.RedisDb)

			return c, err
		},
	}
}
*/
