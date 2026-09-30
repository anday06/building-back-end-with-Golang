package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"task-management-api/internal/config"
	"task-management-api/internal/database"
	"task-management-api/internal/handler"
	"task-management-api/internal/job"
	"task-management-api/internal/middleware"
	"task-management-api/internal/realtime"
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
	backgroundWorker := job.NewWorker(100)
	defer backgroundWorker.Close()
	router := gin.New()
	router.Use(gin.Recovery(), middleware.Logger(), middleware.Metrics(), middleware.RateLimit(), cors.New(cors.Config{AllowOrigins: []string{"*"}, AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, AllowHeaders: []string{"Origin", "Content-Type", "Authorization"}}))
	router.GET("/health", handler.Health)
	router.GET("/metrics", middleware.MetricsHandler())
	websocketHub := realtime.NewHub(cfg.JWTSecret)
	router.GET("/ws", websocketHub.Handle)
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
	comments := handler.CommentHandler{Repo: repository.CommentRepository{DB: db}, Worker: backgroundWorker}
	private.GET("/tasks/:task_id/comments", comments.List)
	private.POST("/comments", comments.Create)
	private.GET("/comments/:id", comments.Get)
	private.PUT("/comments/:id", comments.Update)
	private.DELETE("/comments/:id", comments.Delete)
	server := &http.Server{Addr: ":" + cfg.Port, Handler: router, ReadHeaderTimeout: 5 * time.Second}
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("task management API listening on :%s", cfg.Port)
		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	case <-shutdownSignal.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
	}
}
