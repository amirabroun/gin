package repository

import (
	"gorm.io/gorm"

	"gin/internal/user/entity"
)

type UserRepository interface {
	FindByID(id uint, with ...string) (*entity.User, error)
	FindByMobile(mobile string) (*entity.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(id uint, with ...string) (*entity.User, error) {
	var user entity.User

	query := r.db

	for _, relation := range with {
		query = query.Preload(relation)
	}

	err := query.First(&user, id).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) FindByMobile(mobile string) (*entity.User, error) {
	var user entity.User

	err := r.db.Where("mobile", mobile).First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}
