package models

import (
	"fmt"
	"strings"
)

type User struct {
	ID               uint `gorm:"primaryKey"`
	Contact          string
	Password         string
	SubscriberId     uint
	SubscriberUserId uint
	JointSurveys     []JointSurvey
}

func (u User) String() string {
	var surveys strings.Builder

	if len(u.JointSurveys) == 0 {
		surveys.WriteString("[]")
	} else {
		surveys.WriteString("[\n")
		for i, s := range u.JointSurveys {
			surveys.WriteString(fmt.Sprintf("    %v", s))
			if i < len(u.JointSurveys)-1 {
				surveys.WriteString("\n")
			}
		}
		surveys.WriteString("\n  ]")
	}

	return fmt.Sprintf(
		"User {\n"+
			"  ID:               %d\n"+
			"  Contact:          %q\n"+
			"  SubscriberId:     %d\n"+
			"  SubscriberUserId: %d\n"+
			"  JointSurveys:     %s\n"+
			"}",
		u.ID, u.Contact, u.SubscriberId, u.SubscriberUserId, surveys.String(),
	)
}
