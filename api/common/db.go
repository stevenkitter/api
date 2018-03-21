package common

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
)

const (
	MARIADB_URL = "root:julu666@tcp(115.159.222.199:3306)/julu"
	REDIS_URL   = "redis://root:julu666@115.159.222.199:6379/0?foo=bar&qux=baz"
)

func Mariadb() (db *sql.DB, err error) {
	db, dbErr := sql.Open("mysql", MARIADB_URL)
	return db, dbErr
}

func Redis() (re redis.Conn, err error) {
	c, err := redis.DialURL(REDIS_URL)
	return c, err
}

func RedisString(re redis.Conn, key string) (string, error) {
	s, err := redis.String(re.Do("GET", key))
	return s, err
}
