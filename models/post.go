package models

type Post struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint
	User      User
	Title     string
	Content   string
	Status    string
	ViewCount int
	LikeCount int
}
