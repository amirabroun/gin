package models

import "fmt"

type User struct {
	ID       uint `gorm:"primaryKey"`
	Mobile   string
	Password string
}

func (u User) String() string {
	return fmt.Sprintf(
		"User {\n"+
			"  ID:               %d\n"+
			"  Mobile:          %q\n"+
			"}",
		u.ID, u.Mobile,
	)
}
