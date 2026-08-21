package entity

import "time"

// RoomState tracks the last message a user has read in a room.
// It is used to compute unread counts and read receipts.
type RoomState struct {
	UserID            uint     `gorm:"primaryKey;index" json:"user_id"`
	RoomID            uint     `gorm:"primaryKey;index" json:"room_id"`
	LastReadMessageID *uint    `json:"message_id"`
	LastReadMessage   *Message `gorm:"foreignKey:LastReadMessageID;references:ID"`
	UpdatedAt         time.Time
}

func (RoomState) TableName() string {
	return "room_states"
}
