package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", Health)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(recorder, request)
	require.Equal(t, 200, recorder.Code)
	require.JSONEq(t, `{"status":"ok"}`, recorder.Body.String())
}

func TestPathIDRejectsInvalidValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Params = gin.Params{{Key: "id", Value: "invalid"}}
	_, ok := pathID(context)
	require.False(t, ok)
}

func TestBindRejectsInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest("POST", "/", nil)
	var input struct {
		Name string `json:"name" validate:"required"`
	}
	require.Error(t, bind(context, &input))
}
