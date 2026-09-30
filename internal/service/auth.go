package service

import (
	"errors"
	"gorm.io/gorm"
	"strings"
	"task-management-api/internal/config"
	"task-management-api/internal/models"
	"task-management-api/internal/repository"
	"task-management-api/pkg/password"
	"task-management-api/pkg/token"
)

type AuthService struct {
	DB     *gorm.DB
	Config config.Config
}

type UserService struct{ Repo repository.UserRepository }

func (s UserService) Get(id uint) (models.User, error) { return s.Repo.Get(id) }

func (s UserService) Update(id uint, name, email, rawPassword string) (models.User, error) {
	user, err := s.Repo.Get(id)
	if err != nil {
		return user, err
	}
	user.Name = name
	user.Email = strings.ToLower(email)
	if rawPassword != "" {
		hash, hashErr := password.Hash(rawPassword)
		if hashErr != nil {
			return models.User{}, hashErr
		}
		user.PasswordHash = hash
	}
	return user, s.Repo.Save(&user)
}

func (s UserService) Delete(id uint) error { return s.Repo.Delete(id) }

func (s AuthService) Register(name, email, rawPassword string) (models.User, error) {
	var existing models.User
	if s.DB.Where("email = ?", strings.ToLower(email)).First(&existing).Error == nil {
		return models.User{}, errors.New("email already registered")
	}
	hash, err := password.Hash(rawPassword)
	if err != nil {
		return models.User{}, err
	}
	user := models.User{Name: name, Email: strings.ToLower(email), PasswordHash: hash, Role: "user"}
	return user, s.DB.Create(&user).Error
}
func (s AuthService) Login(email, rawPassword string) (string, models.User, error) {
	var user models.User
	if err := s.DB.Where("email = ?", strings.ToLower(email)).First(&user).Error; err != nil || !password.Compare(user.PasswordHash, rawPassword) {
		return "", user, errors.New("invalid email or password")
	}
	value, err := token.Create(user.ID, user.Role, s.Config.JWTSecret, s.Config.JWTExpiresHour)
	return value, user, err
}
