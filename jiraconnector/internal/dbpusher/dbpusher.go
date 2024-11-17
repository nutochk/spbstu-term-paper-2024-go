package dbpusher

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jiraconnector/internal/dto"
	"github.com/jiraconnector/internal/entities"
	"github.com/jiraconnector/internal/transformer"
)

type DataBase struct {
	Db *pgxpool.Pool
	wm sync.RWMutex
}

func NewDB() (*DataBase, error) {
	conStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		"postgres",
		"elephant",
		"127.0.0.1",
		5432,
		"test",
	)

	pool, err := pgxpool.Connect(context.Background(), conStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %vn", err)
	}
	//defer pool.Close()

	pool.Ping(context.Background())
	if err != nil {
		return nil, err
	}
	return &DataBase{Db: pool}, nil
}

func saveProject(database *DataBase, project *dto.Project) {
	database.Db.QueryRow(context.Background(), "SELECT id from project WHERE title = $1", project.Title).Scan(&project.ID)
	if project.ID == 0 {
		database.wm.Lock()
		database.Db.QueryRow(context.Background(), "INSERT INTO project (title) VALUES($1) RETURNING id", project.Title).Scan(&project.ID)
		database.wm.Unlock()
	}
}

func saveAuthor(database *DataBase, author *dto.Author) {
	database.Db.QueryRow(context.Background(), "SELECT id from author WHERE name = $1", author.Name).Scan(&author.ID)
	if author.ID == 0 {
		database.wm.Lock()
		database.Db.QueryRow(context.Background(), "INSERT INTO author (name) VALUES($1) RETURNING id", author.Name).Scan(&author.ID)
		database.wm.Unlock()
	}
}

func saveStatusChanges(database *DataBase, changes *dto.StatusChanges) {
	database.wm.Lock()
	_, err := database.Db.Exec(context.Background(), "INSERT INTO statuschanges (issueid, authorid, changetime, fromstatus, tostatus) "+
		"VALUES($1, $2, $3, $4, $5)", changes.IssueId, changes.AuthorId, changes.ChangeTime, changes.FromStatus, changes.ToStatus)
	database.wm.Unlock()
	if err != nil {
		log.Fatalf("Unable to save status changes for issue id %d: %v", changes.IssueId, err)
	}
}

func saveIssue(database *DataBase, issue *entities.Issue) {
	issueTrans := transformer.IssueToDTO(issue)
	database.Db.QueryRow(context.Background(), "SELECT id from issue WHERE key = $1", issueTrans.Key).Scan(&issueTrans.ID)
	if issueTrans.ID == 0 {
		project := transformer.ProjectToDTO(&issue.Fields.Project)
		creator := transformer.AuthorToDTO(&issue.Fields.Creator)
		assignee := transformer.AuthorToDTO(&issue.Fields.Assignee)

		saveProject(database, &project)
		saveAuthor(database, &creator)
		saveAuthor(database, &assignee)
		issueTrans.ProjectId = project.ID
		issueTrans.AuthorId = creator.ID
		issueTrans.AssigneeId = assignee.ID

		database.Db.QueryRow(context.Background(), "INSERT INTO issue (projectid, authorid, assigneeid, key, summary, description, type,"+
			" priority, status, createdtime, closedtime, updatedtime, timespent) VALUES($1, $2, $3, $4, $5, $6, $7, $8, "+
			"$9, $10, $11, $12, $13) RETURNING id", issueTrans.ProjectId, issueTrans.AuthorId, issueTrans.AssigneeId, issueTrans.Key,
			issueTrans.Summary, issueTrans.Description, issueTrans.Type, issueTrans.Priority,
			issueTrans.Status, issueTrans.CreatedTime, issueTrans.ClosedTime, issueTrans.UpdatedTime,
			issueTrans.TimeSpent).Scan(&issueTrans.ID)
	}
	for _, history := range issue.StatusChanges.Histories {
		for i := range len(history.Items) {
			changes := transformer.StatusChandesToDTO(&history, i)
			changes.IssueId = issueTrans.ID

			if !isStatusChangesSaved(database, &changes) {
				authorOfChanges := transformer.AuthorToDTO(&history.Author)
				saveAuthor(database, &authorOfChanges)
				changes.AuthorId = authorOfChanges.ID
				saveStatusChanges(database, &changes)
			}
		}
	}
}

func isStatusChangesSaved(database *DataBase, changes *dto.StatusChanges) bool {
	var count int
	database.Db.QueryRow(context.Background(), "IF EXISTS (SELECT TOP 1 1 FROM statuschanges WHERE issueid=$1 AND changetime=$2)", changes.IssueId, changes.ChangeTime).Scan(count)
	return count == 1
}

func Save(database *DataBase, issues *[]entities.Issue) {
	threads := 10
	k := len(*issues) / threads
	o := len(*issues) % threads
	wg := sync.WaitGroup{}

	for i := 0; i < threads; i++ {
		wg.Add(1)
		end := (i + 1) * k
		if i == threads-1 {
			end += o
		}
		go func() {
			for j := i * k; j < end; j++ {
				saveIssue(database, &(*issues)[j])
				wg.Done()
			}
		}()
	}
	wg.Wait()
}

// TODO logger
