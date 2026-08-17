package entity

import "time"

type RoomMember struct {
	RoomID   uint      `gorm:"primaryKey" json:"room_id"`
	UserID   uint      `gorm:"primaryKey" json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
}

func (RoomMember) TableName() string {
	return "room_members"
}