package very_short

import (
	"os"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

//conn data type and it points to sql.DB
type Store struct{
	conn *sql.DB
}
func DatabaseConnection(databaseURL string)(*sql.DB, err){
	return sql.Open("mysql", databaseURL)
}

//recieves database dependency and injects it into the store
func NewStore(conn *sql.DB) *Store{
	return &Store{conn: conn}
}

 
