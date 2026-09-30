package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"task-management-api/internal/config"
	"task-management-api/internal/database"
	"task-management-api/internal/handler"
	"task-management-api/internal/middleware"
	"task-management-api/internal/repository"
	"task-management-api/internal/service"
	"time"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatal(err)
	}
	cache := database.Redis(cfg.RedisURL)
	router := gin.New()
	router.Use(gin.Recovery(), middleware.Logger(), middleware.RateLimit(), cors.New(cors.Config{AllowOrigins: []string{"*"}, AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, AllowHeaders: []string{"Origin", "Content-Type", "Authorization"}}))
	router.GET("/health", handler.Health)
	auth := handler.AuthHandler{Service: service.AuthService{DB: db, Config: cfg}}
	router.POST("/api/v1/auth/register", auth.Register)
	router.POST("/api/v1/auth/login", auth.Login)
	private := router.Group("/api/v1", middleware.Auth(cfg))
	private.GET("/admin/status", middleware.AdminOnly(), handler.AdminStatus)
	users := handler.UserHandler{Service: service.UserService{Repo: repository.UserRepository{DB: db}}}
	private.GET("/users/me", users.Me)
	private.PUT("/users/me", users.UpdateMe)
	private.DELETE("/users/me", users.DeleteMe)
	projects := handler.ProjectHandler{Repo: repository.ProjectRepository{DB: db}}
	private.GET("/projects", projects.List)
	private.POST("/projects", projects.Create)
	private.GET("/projects/:id", projects.Get)
	private.PUT("/projects/:id", projects.Update)
	private.DELETE("/projects/:id", projects.Delete)
	tasks := handler.TaskHandler{Repo: repository.TaskRepository{DB: db, Cache: cache}}
	private.GET("/tasks", tasks.List)
	private.POST("/tasks", tasks.Create)
	private.GET("/tasks/:id", tasks.Get)
	private.PUT("/tasks/:id", tasks.Update)
	private.DELETE("/tasks/:id", tasks.Delete)
	comments := handler.CommentHandler{Repo: repository.CommentRepository{DB: db}}
	private.GET("/tasks/:task_id/comments", comments.List)
	private.POST("/comments", comments.Create)
	private.GET("/comments/:id", comments.Get)
	private.PUT("/comments/:id", comments.Update)
	private.DELETE("/comments/:id", comments.Delete)
	server := &http.Server{Addr: ":" + cfg.Port, Handler: router, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("task management API listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
