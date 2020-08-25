package model

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type Image struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	URL         string
	LabelsStuff pq.StringArray
	MasksLabels pq.StringArray
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func CreateImage(db *gorm.DB, img *Image) error {
	return db.Create(img).Error
}

func AllImagesForProject(db *gorm.DB, projectID uuid.UUID) ([]*Image, error) {
	var i []*Image
	if err := db.Where("project_id = ?", projectID).Find(&i).Error; err != nil {
		return nil, err
	}
	return i, nil
}

func FindImageByID(db *gorm.DB, id uuid.UUID) (*Image, error) {
	var i Image
	if err := db.Where("id = ?", id).Take(&i).Error; err != nil {
		return nil, err
	}
	return &i, nil
}

func DeleteImageByID(db *gorm.DB, imageID uuid.UUID) error {
	return db.Delete(&Image{}, "id = ?", imageID).Error
}
