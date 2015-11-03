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

func (r *Redis) Do(cmd string, args ...interface{}) (interface{}, error) {
	conn := r.pool.Get()
	defer conn.Close()

	return conn.Do(cmd, args...)
}

func (r *Redis) Bool(cmd string, args ...interface{}) (bool, error) {
	return redis.Bool(r.Do(cmd, args...))
}
func (r *Redis) String(cmd string, args ...interface{}) (string, error) {
	return redis.String(r.Do(cmd, args...))
}
func (r *Redis) Strings(cmd string, args ...interface{}) ([]string, error) {
	return redis.Strings(r.Do(cmd, args...))
}
func (r *Redis) StringMap(cmd string, args ...interface{}) (map[string]string, error) {
	return redis.StringMap(r.Do(cmd, args...))
}
