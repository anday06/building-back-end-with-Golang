package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"task-management-api/internal/models"
	"time"
)

type ProjectRepository struct{ DB *gorm.DB }

type UserRepository struct{ DB *gorm.DB }

func (r UserRepository) Get(id uint) (models.User, error) {
	var item models.User
	err := r.DB.First(&item, id).Error
	return item, err
}

func (r UserRepository) Save(item *models.User) error { return r.DB.Save(item).Error }

func (r UserRepository) Delete(id uint) error {
	result := r.DB.Delete(&models.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r ProjectRepository) List(ownerID uint, page, limit int, sort string) ([]models.Project, error) {
	var items []models.Project
	offset := (page - 1) * limit
	err := r.DB.Where("owner_id = ?", ownerID).Preload("Tasks").Order(sort).Offset(offset).Limit(limit).Find(&items).Error
	return items, err
}
func (r ProjectRepository) Get(id, ownerID uint) (models.Project, error) {
	var item models.Project
	err := r.DB.Where("id = ? AND owner_id = ?", id, ownerID).Preload("Tasks").First(&item).Error
	return item, err
}
func (r ProjectRepository) Save(item *models.Project) error { return r.DB.Save(item).Error }
func (r ProjectRepository) Delete(id, ownerID uint) error {
	result := r.DB.Where("id = ? AND owner_id = ?", id, ownerID).Delete(&models.Project{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

type TaskRepository struct {
	DB    *gorm.DB
	Cache *redis.Client
}

type CommentRepository struct{ DB *gorm.DB }

func (r CommentRepository) List(taskID, ownerID uint, page, limit int, sort string) ([]models.Comment, error) {
	var items []models.Comment
	offset := (page - 1) * limit
	err := r.DB.Joins("JOIN tasks ON tasks.id = comments.task_id").Joins("JOIN projects ON projects.id = tasks.project_id").Where("comments.task_id = ? AND projects.owner_id = ?", taskID, ownerID).Order(sort).Offset(offset).Limit(limit).Find(&items).Error
	return items, err
}

func (r CommentRepository) Get(id, ownerID uint) (models.Comment, error) {
	var item models.Comment
	err := r.DB.Joins("JOIN tasks ON tasks.id = comments.task_id").Joins("JOIN projects ON projects.id = tasks.project_id").Where("comments.id = ? AND projects.owner_id = ?", id, ownerID).First(&item).Error
	return item, err
}

func (r CommentRepository) Save(item *models.Comment) error { return r.DB.Save(item).Error }

func (r CommentRepository) Delete(id, authorID uint) error {
	result := r.DB.Where("id = ? AND author_id = ?", id, authorID).Delete(&models.Comment{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r TaskRepository) List(ownerID uint, status string, page, limit int, sort string) ([]models.Task, error) {
	var items []models.Task
	query := r.DB.Where("project_id IN (SELECT id FROM projects WHERE owner_id = ?)", ownerID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	offset := (page - 1) * limit
	err := query.Order(sort).Offset(offset).Limit(limit).Find(&items).Error
	return items, err
}
func (r TaskRepository) Get(id, ownerID uint) (models.Task, error) {
	var item models.Task
	err := r.DB.Where("tasks.id = ? AND projects.owner_id = ?", id, ownerID).Joins("JOIN projects ON projects.id = tasks.project_id").First(&item).Error
	return item, err
}
func (r TaskRepository) Save(item *models.Task) error { return r.DB.Save(item).Error }
func (r TaskRepository) Delete(id, ownerID uint) error {
	result := r.DB.Where("id = ? AND project_id IN (SELECT id FROM projects WHERE owner_id = ?)", id, ownerID).Delete(&models.Task{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (r TaskRepository) Invalidate(ctx context.Context, ownerID uint) {
	if r.Cache != nil {
		keys, err := r.Cache.Keys(ctx, fmt.Sprintf("tasks:%d:*", ownerID)).Result()
		if err == nil && len(keys) > 0 {
			_ = r.Cache.Del(ctx, keys...).Err()
		}
	}
}
func (r TaskRepository) CachedList(ctx context.Context, ownerID uint, status string, page, limit int, sort string) ([]models.Task, error) {
	key := fmt.Sprintf("tasks:%d:%s:%d:%d:%s", ownerID, status, page, limit, sort)
	if r.Cache != nil {
		if raw, err := r.Cache.Get(ctx, key).Result(); err == nil {
			var items []models.Task
			if json.Unmarshal([]byte(raw), &items) == nil {
				return items, nil
			}
		}
	}
	items, err := r.List(ownerID, status, page, limit, sort)
	if err != nil {
		return nil, err
	}
	if r.Cache != nil {
		if raw, marshalErr := json.Marshal(items); marshalErr == nil {
			_ = r.Cache.Set(ctx, key, raw, time.Minute).Err()
		}
	}
	return items, nil
}
