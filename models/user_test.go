package models

import (
	"testing"

	"gin/database"
	"gin/testhelpers"
)

func TestUserHasJointSurveys(t *testing.T) {
	var user User

	db := database.InitDB()
	err := db.Preload("JointSurveys").First(&user, 1).Error

	testhelpers.AssertNoError(t, err, user)
}
