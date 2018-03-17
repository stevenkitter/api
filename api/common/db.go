package common

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func Mariadb() (db *sql.DB, err error) {
	db, dbErr := sql.Open("mysql", "root:julu666@tcp(115.159.222.199:3306)/julu")
	return db, dbErr
}
