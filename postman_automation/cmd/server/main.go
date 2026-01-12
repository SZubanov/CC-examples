package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/postman-automation/task-manager/internal/handler"
	"github.com/postman-automation/task-manager/internal/middleware"
	"github.com/postman-automation/task-manager/internal/service"
	"github.com/postman-automation/task-manager/internal/storage"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	store := storage.New()
	authService := service.NewAuthService(store)
	taskService := service.NewTaskService(store)

	authHandler := handler.NewAuthHandler(authService)
	taskHandler := handler.NewTaskHandler(taskService)
	healthHandler := handler.NewHealthHandler()

	authMiddleware := middleware.NewAuthMiddleware(authService)

	r := mux.NewRouter()
	r.Use(middleware.LoggingMiddleware)

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/health", healthHandler.Health).Methods("GET")

	authRouter := api.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/register", authHandler.Register).Methods("POST")
	authRouter.HandleFunc("/login", authHandler.Login).Methods("POST")

	taskRouter := api.PathPrefix("/tasks").Subrouter()
	taskRouter.Use(authMiddleware.Authenticate)
	taskRouter.HandleFunc("", taskHandler.CreateTask).Methods("POST")
	taskRouter.HandleFunc("", taskHandler.GetTasks).Methods("GET")
	taskRouter.HandleFunc("/{id}", taskHandler.GetTask).Methods("GET")
	taskRouter.HandleFunc("/{id}", taskHandler.UpdateTask).Methods("PUT")
	taskRouter.HandleFunc("/{id}", taskHandler.DeleteTask).Methods("DELETE")
	taskRouter.HandleFunc("/{id}/status", taskHandler.UpdateStatus).Methods("PATCH")

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
