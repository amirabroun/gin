package models

import (
	"gin/database"
	"gin/testhelpers"
	"testing"
)

func TestJointSurveyHasUser(t *testing.T) {
	var jointSurvey JointSurvey

	db := database.InitDB()
	err := db.Preload("User").First(&jointSurvey, 232169).Error

	testhelpers.AssertNoError(t, err, jointSurvey)
}
