package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"task-management-api/internal/job"
	"task-management-api/internal/models"
	"task-management-api/internal/repository"
	"task-management-api/pkg/response"
)

var validate = validator.New()

func bind(c *gin.Context, value any) error {
	if err := c.ShouldBindJSON(value); err != nil {
		response.Error(c, 400, "invalid JSON body")
		return err
	}
	if err := validate.Struct(value); err != nil {
		response.Error(c, 400, "validation failed")
		return err
	}
	return nil
}
func currentUser(c *gin.Context) uint { value, _ := c.Get("user_id"); id, _ := value.(uint); return id }
func pathID(c *gin.Context) (uint, bool) {
	return pathParam(c, "id")
}
func pathParam(c *gin.Context, name string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 32)
	if err != nil {
		response.Error(c, 400, "invalid id")
		return 0, false
	}
	return uint(id), true
}

func listParams(c *gin.Context) (int, int, string) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	sort := "created_at DESC"
	if c.Query("sort") == "oldest" {
		sort = "created_at ASC"
	}
	return page, limit, sort
}

type ProjectHandler struct{ Repo repository.ProjectRepository }

// @Summary List projects
// @Description Get paginated list of user's projects
// @Tags projects
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param sort query string false "Sort order: created_at DESC or ASC" default(created_at DESC)
// @Success 200 {array} models.Project
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /projects [get]
func (h ProjectHandler) List(c *gin.Context) {
	page, limit, sort := listParams(c)
	items, err := h.Repo.List(currentUser(c), page, limit, sort)
	if err != nil {
		response.Error(c, 500, "could not list projects")
		return
	}
	response.Success(c, 200, items)
}

// @Summary Get project by ID
// @Description Get a single project by ID
// @Tags projects
// @Security BearerAuth
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {object} models.Project
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /projects/{id} [get]
func (h ProjectHandler) Get(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	item, err := h.Repo.Get(id, currentUser(c))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, 404, "project not found")
		return
	}
	if err != nil {
		response.Error(c, 500, "could not get project")
		return
	}
	response.Success(c, 200, item)
}

// @Summary Create project
// @Description Create a new project
// @Tags projects
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ProjectCreateRequest true "Project data"
// @Success 201 {object} models.Project
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /projects [post]
func (h ProjectHandler) Create(c *gin.Context) {
	var input struct {
		Name        string `json:"name" validate:"required,min=2,max=120"`
		Description string `json:"description" validate:"max=1000"`
	}
	if bind(c, &input) != nil {
		return
	}
	item := models.Project{Name: input.Name, Description: input.Description, OwnerID: currentUser(c)}
	if err := h.Repo.Save(&item); err != nil {
		response.Error(c, 500, "could not create project")
		return
	}
	response.Success(c, 201, item)
}

// @Summary Update project
// @Description Update an existing project
// @Tags projects
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param request body ProjectCreateRequest true "Project data"
// @Success 200 {object} models.Project
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /projects/{id} [put]
func (h ProjectHandler) Update(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	item, err := h.Repo.Get(id, currentUser(c))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, 404, "project not found")
		return
	}
	if err != nil {
		response.Error(c, 500, "could not get project")
		return
	}
	var input struct {
		Name        string `json:"name" validate:"required,min=2,max=120"`
		Description string `json:"description" validate:"max=1000"`
	}
	if bind(c, &input) != nil {
		return
	}
	item.Name, item.Description = input.Name, input.Description
	if err := h.Repo.Save(&item); err != nil {
		response.Error(c, 500, "could not update project")
		return
	}
	response.Success(c, 200, item)
}

// @Summary Delete project
// @Description Delete a project
// @Tags projects
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /projects/{id} [delete]
func (h ProjectHandler) Delete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.Repo.Delete(id, currentUser(c)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, 404, "project not found")
			return
		}
		response.Error(c, 500, "could not delete project")
		return
	}
	c.Status(204)
}

type TaskHandler struct{ Repo repository.TaskRepository }

// @Summary List tasks
// @Description Get paginated list of tasks with optional status filter
// @Tags tasks
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param sort query string false "Sort order" default(created_at DESC)
// @Param status query string false "Filter by status" Enums(todo, in_progress, done)
// @Success 200 {array} models.Task
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /tasks [get]
func (h TaskHandler) List(c *gin.Context) {
	page, limit, sort := listParams(c)
	items, err := h.Repo.CachedList(c, currentUser(c), c.Query("status"), page, limit, sort)
	if err != nil {
		response.Error(c, 500, "could not list tasks")
		return
	}
	response.Success(c, 200, items)
}

// @Summary Get task by ID
// @Description Get a single task by ID
// @Tags tasks
// @Security BearerAuth
// @Produce json
// @Param id path int true "Task ID"
// @Success 200 {object} models.Task
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /tasks/{id} [get]
func (h TaskHandler) Get(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	item, err := h.Repo.Get(id, currentUser(c))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, 404, "task not found")
		return
	}
	if err != nil {
		response.Error(c, 500, "could not get task")
		return
	}
	response.Success(c, 200, item)
}

// @Summary Create task
// @Description Create a new task
// @Tags tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body TaskCreateRequest true "Task data"
// @Success 201 {object} models.Task
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /tasks [post]
func (h TaskHandler) Create(c *gin.Context) {
	var input struct {
		Title       string `json:"title" validate:"required,min=2,max=200"`
		Description string `json:"description" validate:"max=2000"`
		Status      string `json:"status" validate:"omitempty,oneof=todo in_progress done"`
		Priority    string `json:"priority" validate:"omitempty,oneof=low medium high"`
		ProjectID   uint   `json:"project_id" validate:"required"`
	}
	if bind(c, &input) != nil {
		return
	}
	item := models.Task{Title: input.Title, Description: input.Description, Status: input.Status, Priority: input.Priority, ProjectID: input.ProjectID}
	if item.Status == "" {
		item.Status = "todo"
	}
	if item.Priority == "" {
		item.Priority = "medium"
	}
	if err := h.Repo.DB.Where("id = ? AND owner_id = ?", item.ProjectID, currentUser(c)).First(&models.Project{}).Error; err != nil {
		response.Error(c, 404, "project not found")
		return
	}
	if err := h.Repo.Save(&item); err != nil {
		response.Error(c, 500, "could not create task")
		return
	}
	h.Repo.Invalidate(c, currentUser(c))
	response.Success(c, 201, item)
}

// @Summary Update task
// @Description Update an existing task
// @Tags tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param request body TaskUpdateRequest true "Task data"
// @Success 200 {object} models.Task
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /tasks/{id} [put]
func (h TaskHandler) Update(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	item, err := h.Repo.Get(id, currentUser(c))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, 404, "task not found")
		return
	}
	if err != nil {
		response.Error(c, 500, "could not get task")
		return
	}
	var input struct {
		Title       string `json:"title" validate:"required,min=2,max=200"`
		Description string `json:"description" validate:"max=2000"`
		Status      string `json:"status" validate:"required,oneof=todo in_progress done"`
		Priority    string `json:"priority" validate:"required,oneof=low medium high"`
	}
	if bind(c, &input) != nil {
		return
	}
	item.Title, item.Description, item.Status, item.Priority = input.Title, input.Description, input.Status, input.Priority
	if err := h.Repo.Save(&item); err != nil {
		response.Error(c, 500, "could not update task")
		return
	}
	h.Repo.Invalidate(c, currentUser(c))
	response.Success(c, 200, item)
}

// @Summary Delete task
// @Description Delete a task
// @Tags tasks
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /tasks/{id} [delete]
func (h TaskHandler) Delete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.Repo.Delete(id, currentUser(c)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, 404, "task not found")
			return
		}
		response.Error(c, 500, "could not delete task")
		return
	}
	h.Repo.Invalidate(c, currentUser(c))
	c.Status(204)
}

type CommentHandler struct {
	Repo   repository.CommentRepository
	Worker *job.Worker
}

// @Summary List comments for a task
// @Description Get paginated list of comments for a task
// @Tags comments
// @Security BearerAuth
// @Produce json
// @Param id path int true "Task ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param sort query string false "Sort order" default(created_at DESC)
// @Success 200 {array} models.Comment
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /tasks/{id}/comments [get]
func (h CommentHandler) List(c *gin.Context) {
	taskID, ok := pathID(c)
	if !ok {
		return
	}
	page, limit, sort := listParams(c)
	items, err := h.Repo.List(taskID, currentUser(c), page, limit, sort)
	if err != nil {
		response.Error(c, 500, "could not list comments")
		return
	}
	response.Success(c, 200, items)
}

// @Summary Get comment by ID
// @Description Get a single comment by ID
// @Tags comments
// @Security BearerAuth
// @Produce json
// @Param id path int true "Comment ID"
// @Success 200 {object} models.Comment
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /comments/{id} [get]
func (h CommentHandler) Get(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	item, err := h.Repo.Get(id, currentUser(c))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, 404, "comment not found")
		return
	}
	if err != nil {
		response.Error(c, 500, "could not get comment")
		return
	}
	response.Success(c, 200, item)
}

// @Summary Create comment
// @Description Create a new comment on a task
// @Tags comments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CommentCreateRequest true "Comment data"
// @Success 201 {object} models.Comment
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /comments [post]
func (h CommentHandler) Create(c *gin.Context) {
	var input struct {
		Body   string `json:"body" validate:"required,min=1,max=2000"`
		TaskID uint   `json:"task_id" validate:"required"`
	}
	if bind(c, &input) != nil {
		return
	}
	var task models.Task
	if err := h.Repo.DB.Joins("JOIN projects ON projects.id = tasks.project_id").Where("tasks.id = ? AND projects.owner_id = ?", input.TaskID, currentUser(c)).First(&task).Error; err != nil {
		response.Error(c, 404, "task not found")
		return
	}
	item := models.Comment{Body: input.Body, TaskID: input.TaskID, AuthorID: currentUser(c)}
	if err := h.Repo.Save(&item); err != nil {
		response.Error(c, 500, "could not create comment")
		return
	}
	if h.Worker != nil {
		h.Worker.Enqueue(job.Notification{TaskID: item.TaskID, AuthorID: item.AuthorID, Body: item.Body})
	}
	response.Success(c, 201, item)
}

// @Summary Update comment
// @Description Update a comment (only author)
// @Tags comments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Comment ID"
// @Param request body CommentUpdateRequest true "Comment data"
// @Success 200 {object} models.Comment
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /comments/{id} [put]
func (h CommentHandler) Update(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	item, err := h.Repo.Get(id, currentUser(c))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, 404, "comment not found")
		return
	}
	if err != nil {
		response.Error(c, 500, "could not get comment")
		return
	}
	if item.AuthorID != currentUser(c) {
		response.Error(c, 403, "only the author can update this comment")
		return
	}
	var input struct {
		Body string `json:"body" validate:"required,min=1,max=2000"`
	}
	if bind(c, &input) != nil {
		return
	}
	item.Body = input.Body
	if err := h.Repo.Save(&item); err != nil {
		response.Error(c, 500, "could not update comment")
		return
	}
	response.Success(c, 200, item)
}

// @Summary Delete comment
// @Description Delete a comment (only author)
// @Tags comments
// @Security BearerAuth
// @Param id path int true "Comment ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /comments/{id} [delete]
func (h CommentHandler) Delete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.Repo.Delete(id, currentUser(c)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, 404, "comment not found or not owned by user")
			return
		}
		response.Error(c, 500, "could not delete comment")
		return
	}
	c.Status(204)
}

// @Summary Health check
// @Description Check API health
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func Health(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) }

// @Summary Admin status
// @Description Admin only endpoint
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /admin/status [get]
func AdminStatus(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{"message": "admin access granted"})
}

// Request/Response types for Swagger
type ProjectCreateRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=120"`
	Description string `json:"description" validate:"max=1000"`
}

type TaskCreateRequest struct {
	Title       string `json:"title" validate:"required,min=2,max=200"`
	Description string `json:"description" validate:"max=2000"`
	Status      string `json:"status" validate:"omitempty,oneof=todo in_progress done"`
	Priority    string `json:"priority" validate:"omitempty,oneof=low medium high"`
	ProjectID   uint   `json:"project_id" validate:"required"`
}

type TaskUpdateRequest struct {
	Title       string `json:"title" validate:"required,min=2,max=200"`
	Description string `json:"description" validate:"max=2000"`
	Status      string `json:"status" validate:"required,oneof=todo in_progress done"`
	Priority    string `json:"priority" validate:"required,oneof=low medium high"`
}

type CommentCreateRequest struct {
	Body   string `json:"body" validate:"required,min=1,max=2000"`
	TaskID uint   `json:"task_id" validate:"required"`
}

type CommentUpdateRequest struct {
	Body string `json:"body" validate:"required,min=1,max=2000"`
}
