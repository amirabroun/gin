package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"gin/internal/user/entity"
	"gin/internal/user/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

// UserService holds the domain/business logic of the user bounded context.
type UserService struct {
	repo repository.UserRepository
}

func New(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetByID(id uint, with ...string) (*entity.User, error) {
	return s.repo.FindByID(id, with...)
}

func (s *UserService) List() ([]entity.User, error) {
	return s.repo.List()
}

func (s *UserService) StoreUserPost(post entity.Post) error {
	return s.repo.StorePost(post)
}

func (s *UserService) Login(mobile, password string) (*entity.User, error) {
	user, err := s.repo.FindByMobile(mobile)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (s *UserService) Register(mobile, password string) (*entity.User, error) {
	if _, err := s.repo.FindByMobile(mobile); err == nil {
		return nil, ErrUserAlreadyExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Mobile:   mobile,
		Password: string(hashed),
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}
