package next

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

type Redis struct {
	pool *redis.Pool
}

func NewRedis() *Redis {
	return &Redis{}
}

func (r *Redis) Pool(server string, pwd ...interface{}) {
	r.pool = &redis.Pool{
		MaxIdle:     64,
		IdleTimeout: 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if len(pwd) > 0 {
				if _, err := c.Do("AUTH", pwd[0]); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func (r *Redis) Do(cmd string, args ...interface{}) (replay interface{}, err error) {
	conn := r.pool.Get()
	defer conn.Close()

	return conn.Do(cmd, args...)
}
