package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"task-management-api/internal/models"
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
type authResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// @Summary Register a new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body registerRequest true "Register request"
// @Success 201 {object} models.User
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /auth/register [post]
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

// @Summary Login user
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body loginRequest true "Login request"
// @Success 200 {object} authResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /auth/login [post]
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
