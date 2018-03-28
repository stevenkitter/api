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
func RedisGETString(key string) (string, error) {
	redis, err := Redis()
	defer redis.Close()
	if err != nil {
		return "", err
	}
	s, err := RedisString(redis, Md5(key))
	return s, err
}

func RedisSaveString(key string, value string) error {
	redis, err := Redis()
	defer redis.Close()
	if err != nil {
		return err
	}
	md5_key := Md5(key)
	_, err = redis.Do("SET", md5_key, value) //"EX", "5"
	if err != nil {
		return err
	}
	return nil
}

func RedisSaveStringEx(key string, value string, ex string) error {
	redis, err := Redis()
	defer redis.Close()
	if err != nil {
		return err
	}
	md5_key := Md5(key)
	_, err = redis.Do("SET", md5_key, value, "EX", ex)
	if err != nil {
		return err
	}
	return nil
}
