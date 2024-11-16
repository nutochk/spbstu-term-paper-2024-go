package transformer

import (
	"strconv"
	"time"

	"github.com/jiraconnector/internal/dto"
	"github.com/jiraconnector/internal/entities"
)

func AuthorToDTO(creator *entities.Creator) dto.Author {
	return dto.Author{
		Name: creator.Name,
	}
}

func ProjectToDTO(project *entities.Project) dto.Project {
	return dto.Project{
		Title: project.Name,
	}
}

func IssueToDTO(issue *entities.Issue) dto.Issue {
	createdTime, _ := time.Parse("2006-01-02T15:04:05.999-0700", issue.Fields.CreatedTime)
	updatedTime, _ := time.Parse("2006-01-02T15:04:05.999-0700", issue.Fields.UpdatedTime)
	closedTime, _ := time.Parse("2006-01-02T15:04:05.999-0700", issue.Fields.ClosedTime)
	timeSpent, _ := strconv.Atoi(issue.Fields.TimeSpent)
	return dto.Issue{
		Key:         issue.Key,
		CreatedTime: createdTime,
		UpdatedTime: updatedTime,
		ClosedTime:  closedTime,
		TimeSpent:   timeSpent,
		Summary:     issue.Fields.Summary,
		Description: issue.Fields.Description,
		Priority:    issue.Fields.Priority.Name,
		Status:      issue.Fields.Status.Name,
		Type:        issue.Fields.Type.Name,
	}
}

func StatusChandesToDTO(history *entities.History, i int) dto.StatusChanges {
	changeTime, _ := time.Parse("2006-01-02T15:04:05.999-0700", history.ChangeTime)
	return dto.StatusChanges{
		ChangeTime: changeTime,
		FromStatus: history.Items[i].FromString,
		ToStatus:   history.Items[i].ToString,
	}
}
