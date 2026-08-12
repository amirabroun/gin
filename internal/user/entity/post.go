package entity

type Post struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Status    string `gorm:"default:draft" json:"status"`
	ViewCount int    `json:"view_count"`
	LikeCount int    `json:"like_count"`
	UserID    uint   `json:"user_id"`
	User      User   `json:"user"`
}
