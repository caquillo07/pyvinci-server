package model

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type Project struct {
	ID        uuid.UUID
	UserID    uuid.UUID `gorm:"column:user_record"`
	Name      string
	Keywords  pq.StringArray
	CreatedAt time.Time
	UpdatedAt time.Time
}

func CreateProject(db *gorm.DB, project *Project) error {
	return db.Create(&project).Error
}

func AllProjectsForUser(db *gorm.DB, userID uuid.UUID) ([]*Project, error) {
	var p []*Project
	if err := db.Where("user_record = ?", userID).Find(&p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

func FindProjectByID(db *gorm.DB, projectID uuid.UUID) (*Project, error) {
	var p Project
	if err := db.Where("id = ?", projectID).Take(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func DeleteProjectByID(db *gorm.DB, projectID uuid.UUID) error {
	return db.Delete(&Project{}, "id = ?", projectID).Error
}

func (p *Project) Update(db *gorm.DB) error {
	return db.Save(p).Error
}
