package model

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
)

type Job struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (*Job) TableName() string {
	return "jobs"
}

func CreateNewJob(db *gorm.DB, projectID uuid.UUID) (*Job, error) {
	j := Job{
		ProjectID: projectID,
		Status:    "PENDING_LABELS",
	}
	if err := db.Create(&j).Error; err != nil {
		return nil, err
	}
	return &j, nil
}

func FindJobForProject(db *gorm.DB, projectID uuid.UUID) (*Job, error) {
	var j Job
	if err := db.Where("project_id = ?", projectID).Take(&j).Error; err != nil {
		return nil, err
	}
	return &j, nil
}
