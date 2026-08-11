package entity

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Mobile   string `json:"mobile"`
	Password string `json:"-"`
	Posts    []Post `json:"posts,omitempty"`
}
