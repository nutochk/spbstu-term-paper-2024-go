package dbpusher

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var (
	host     = "127.0.0.1"
	port     = 5432
	user     = "postgres"
	password = "elephant"
	db       = "test"
)

func DbPusher(projectTitle string) {
	dataSourceName := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, db)
	database, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		panic(err)
	}
	var projectId int
	database.QueryRow("INSERT INTO \"projects\" (title) VALUES($1) RETURNING id", projectTitle).Scan(&projectId)
	//database.QueryRow("SELECT id FROM \"projects\" WHERE title=$1", projectTitle).Scan(&projectId)
	fmt.Println(projectId)

}
