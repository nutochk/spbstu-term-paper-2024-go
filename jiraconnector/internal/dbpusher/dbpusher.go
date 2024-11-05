package dbpusher

import (
	"database/sql"
	"fmt"
	"github.com/jiraconnector/internal/dto"
	"github.com/jiraconnector/internal/entities"
	"github.com/jiraconnector/internal/transformer"
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

func saveProject(database *DataBase, project *dto.Project) {
	database.Db.QueryRow("SELECT id from project WHERE title = $1", project.Title).Scan(&project.ID)
	if project.ID == 0 {
		database.Db.QueryRow("INSERT INTO project (title) VALUES($1) RETURNING id", project.Title).Scan(&project.ID)
	}
}

func saveAuthor(database *DataBase, author *dto.Author) {
	database.Db.QueryRow("SELECT id from author WHERE name = $1", author.Name).Scan(&author.ID)
	if author.ID == 0 {
		database.Db.QueryRow("INSERT INTO author (name) VALUES($1) RETURNING id", author.Name).Scan(&author.ID)
	}
}

// TODO проверка записи на существование????
func SaveIssue(database *DataBase, issue *entities.Issue) {
	project := transformer.ProjectToDTO(&issue.Fields.Project)
	creator := transformer.AuthorToDTO(&issue.Fields.Creator)
	assignee := transformer.AuthorToDTO(&issue.Fields.Assignee)
	issueTrans := transformer.IssueToDTO(issue)
	saveProject(database, &project)
	saveAuthor(database, &creator)
	saveAuthor(database, &assignee)
	issueTrans.ProjectId = project.ID
	issueTrans.AuthorId = creator.ID
	issueTrans.AssigneeId = assignee.ID
	result, err := database.Db.Exec("INSERT INTO issue (projectid, authorid, assigneeid, key, summary, description, type,"+
		" priority, status, createdtime, closedtime, updatetime, timespent) VALUES($1, $2, $3, $4, $5, $6, $7, $8, "+
		"$9, $10, $11, $12, $13) RETURNING id", issueTrans.ProjectId, issueTrans.AuthorId, issueTrans.AssigneeId, issueTrans.Key,
		issueTrans.Summary, issueTrans.Description, issueTrans.Type, issueTrans.Priority,
		issueTrans.Status, issueTrans.CreatedTime, issueTrans.ClosedTime, issueTrans.UpdatedTime,
		issueTrans.TimeSpent)
	fmt.Errorf("\t", err)
	id, err := result.LastInsertId()
	fmt.Println(id)
}
