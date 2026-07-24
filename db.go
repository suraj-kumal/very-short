package very_short

import (
	"os"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func DatabaseConnection(databaseURL string)(*sql.DB, err){
	return sql.Open("mysql", databaseURL)
}

