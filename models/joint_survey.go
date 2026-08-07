package models

import "fmt"

type JointSurvey struct {
	ID     uint `gorm:"primaryKey"`
	UserID uint
	User   User
	Name   string
	Gender int
}

func (s JointSurvey) String() string {
	return fmt.Sprintf(
		"JointSurvey{ID: %d, Name: %q, Gender: %d}",
		s.ID, s.Name, s.Gender,
	)
}
