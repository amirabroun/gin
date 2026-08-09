package repository

import (
	"gorm.io/gorm"

	"gin/models"
)

type UserRepository interface {
	FindByID(id uint) (*models.User, error)
	FindByMobile(mobile string) (*models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByMobile(mobile string) (*models.User, error) {
	var user models.User

	err := r.db.Where("mobile", mobile).First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}
