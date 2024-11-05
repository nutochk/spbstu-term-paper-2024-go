package dbpusher

import (
	"database/sql"
	"fmt"
	"github.com/jiraconnector/internal/dto"
	_ "github.com/lib/pq"
)

type DataBase struct {
	Db *sql.DB
}

func NewDB() (*DataBase, error) {
	conStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		"postgres",
		"elephant",
		"127.0.0.1",
		5432,
		"test",
	)
	db, err := sql.Open("postgres", conStr)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &DataBase{Db: db}, nil
}

func SaveProject(database *DataBase, project *dto.Project) (projectId int) {
	database.Db.QueryRow("SELECT id from project WHERE title = $1", project.Title).Scan(&projectId)
	if projectId == 0 {
		database.Db.QueryRow("INSERT INTO project (title) VALUES($1) RETURNING id", project.Title).Scan(&projectId)
	}
	return projectId
}
