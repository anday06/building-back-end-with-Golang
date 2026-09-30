package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"task-management-api/internal/config"
	"task-management-api/pkg/token"
)

func TestAuthRejectsMissingToken(t *testing.T) {
	router := gin.New()
	router.GET("/private", Auth(config.Config{JWTSecret: "secret"}), func(c *gin.Context) { c.Status(http.StatusOK) })
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/private", nil))
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestAuthAcceptsValidToken(t *testing.T) {
	value, err := token.Create(7, "user", "secret", 1)
	require.NoError(t, err)
	router := gin.New()
	router.GET("/private", Auth(config.Config{JWTSecret: "secret"}), func(c *gin.Context) { c.Status(http.StatusOK) })
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.Header.Set("Authorization", "Bearer "+value)
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestAdminOnlyRejectsRegularUser(t *testing.T) {
	router := gin.New()
	router.GET("/admin", func(c *gin.Context) {
		c.Set("role", "user")
		c.Next()
	}, AdminOnly(), func(c *gin.Context) { c.Status(http.StatusOK) })
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/admin", nil))
	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestAdminOnlyAllowsAdmin(t *testing.T) {
	router := gin.New()
	router.GET("/admin", func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	}, AdminOnly(), func(c *gin.Context) { c.Status(http.StatusOK) })
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/admin", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestAuthRejectsWrongSecret(t *testing.T) {
	value, err := token.Create(7, "user", "secret", 1)
	require.NoError(t, err)
	router := gin.New()
	router.GET("/private", Auth(config.Config{JWTSecret: "wrong-secret"}), func(c *gin.Context) { c.Status(http.StatusOK) })
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.Header.Set("Authorization", "Bearer "+value)
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestCORSAddsAllowOriginHeader(t *testing.T) {
	router := gin.New()
	router.Use(cors.New(cors.Config{AllowOrigins: []string{"https://example.com"}, AllowMethods: []string{"GET"}}))
	router.GET("/public", func(c *gin.Context) { c.Status(http.StatusOK) })
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/public", nil)
	request.Host = "api.localhost"
	request.Header.Set("Origin", "https://example.com")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	router.ServeHTTP(recorder, request)
	require.Equal(t, "https://example.com", recorder.Header().Get("Access-Control-Allow-Origin"))
}
