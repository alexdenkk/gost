package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Lab struct {
	gorm.Model

	ID       uuid.UUID `json:"id" gorm:"primaryKey"`
	UserID   uuid.UUID `json:"user_id"`
	Filename string    `json:"filename"`

	Date string `json:"date"`

	Department     string `json:"department"`
	Professor      string `json:"professor"`
	ProfessorGrade string `json:"professor_grade"`
	Course         string `json:"course"`
	Group          string `json:"group"`
	Student        string `json:"student"`

	LabNumber string `json:"lab_number"`
	LabTitle  string `json:"lab_title"`

	Text   string  `json:"text" gorm:"-"`
	Images []Image `json:"images" gorm:"-"`
}

type Image struct {
	Description string `json:"description"`
	Base64      string `json:"base_64"`
}

func (lab *Lab) Validate() error {
	return nil
}
