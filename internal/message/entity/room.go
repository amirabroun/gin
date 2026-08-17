package entity

import "time"

const (
	RoomTypeDirect = "direct"
	RoomTypeGroup  = "group"
)

type Room struct {
	ID        uint         `gorm:"primaryKey" json:"id"`
	Type      string       `gorm:"not null" json:"type"`
	Name      *string      `json:"name,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	Members   []RoomMember `gorm:"foreignKey:RoomID" json:"members,omitempty"`
}
