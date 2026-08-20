package entity

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `json:"name"`
	Mobile   string `json:"mobile"`
	Password string `json:"-"`
	Posts    []Post `json:"posts,omitempty"`
}
