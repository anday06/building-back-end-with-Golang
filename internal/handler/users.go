package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"task-management-api/internal/service"
	"task-management-api/pkg/response"
)

type UserHandler struct{ Service service.UserService }

func (h UserHandler) Me(c *gin.Context) {
	user, err := h.Service.Get(currentUser(c))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "could not get user")
		return
	}
	response.Success(c, http.StatusOK, user)
}

func (h UserHandler) UpdateMe(c *gin.Context) {
	var input struct {
		Name     string `json:"name" validate:"required,min=2,max=100"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"omitempty,min=8"`
	}
	if bind(c, &input) != nil {
		return
	}
	user, err := h.Service.Update(currentUser(c), input.Name, input.Email, input.Password)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusConflict, "could not update user")
		return
	}
	response.Success(c, http.StatusOK, user)
}

func (h UserHandler) DeleteMe(c *gin.Context) {
	err := h.Service.Delete(currentUser(c))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "could not delete user")
		return
	}
	c.Status(http.StatusNoContent)
}
