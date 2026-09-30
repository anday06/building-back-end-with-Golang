package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `json:"name" validate:"required,min=2,max=100"`
	Email        string         `gorm:"uniqueIndex;not null" json:"email" validate:"required,email"`
	PasswordHash string         `json:"-"`
	Role         string         `gorm:"default:user" json:"role"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-"`
	Projects     []Project      `gorm:"foreignKey:OwnerID" json:"-"`
}

type Project struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `json:"name" validate:"required,min=2,max=120"`
	Description string         `json:"description" validate:"max=1000"`
	OwnerID     uint           `json:"owner_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-"`
	Tasks       []Task         `json:"tasks,omitempty"`
}

type Task struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `json:"title" validate:"required,min=2,max=200"`
	Description string         `json:"description" validate:"max=2000"`
	Status      string         `gorm:"default:todo" json:"status" validate:"omitempty,oneof=todo in_progress done"`
	Priority    string         `gorm:"default:medium" json:"priority" validate:"omitempty,oneof=low medium high"`
	ProjectID   uint           `json:"project_id" validate:"required"`
	AssigneeID  *uint          `json:"assignee_id,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-"`
}

type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Body      string         `json:"body" validate:"required,min=1,max=2000"`
	TaskID    uint           `json:"task_id"`
	AuthorID  uint           `json:"author_id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-"`
}
