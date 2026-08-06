package models


type User struct {
	ID               uint `gorm:"primaryKey"`
	Contact          string
	SubscriberId     uint
	SubscriberUserId uint
}
