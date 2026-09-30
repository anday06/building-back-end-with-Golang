package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"task-management-api/internal/service"
	"task-management-api/pkg/response"
)

type AuthHandler struct{ Service service.AuthService }
type registerRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}
type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := bind(c, &req); err != nil {
		return
	}
	user, err := h.Service.Register(req.Name, req.Email, req.Password)
	if err != nil {
		response.Error(c, 409, err.Error())
		return
	}
	response.Success(c, http.StatusCreated, user)
}
func (h AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := bind(c, &req); err != nil {
		return
	}
	value, user, err := h.Service.Login(req.Email, req.Password)
	if err != nil {
		response.Error(c, 401, err.Error())
		return
	}
	response.Success(c, http.StatusOK, gin.H{"token": value, "user": user})
}
