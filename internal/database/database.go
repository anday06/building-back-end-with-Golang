package database

import (
	"context"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"task-management-api/internal/models"
)

func Connect(url string) (*gorm.DB, error) { return gorm.Open(postgres.Open(url), &gorm.Config{}) }
func Redis(url string) *redis.Client       { return redis.NewClient(&redis.Options{Addr: url}) }
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{}, &models.Project{}, &models.Task{}, &models.Comment{})
}
func PingRedis(client *redis.Client) error { return client.Ping(context.Background()).Err() }
