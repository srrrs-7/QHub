package main

import (
	"api/src/infra/rds/organization_repository"
	"api/src/infra/rds/project_repository"
	"api/src/infra/rds/prompt_repository"
	"api/src/infra/rds/task_repository"
	"api/src/routes"
	"api/src/routes/organizations"
	"api/src/routes/projects"
	"api/src/routes/prompts"
	"api/src/routes/tasks"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"utils/db"
	dbq "utils/db/db"
	"utils/env"
	"utils/logger"
)

func init() {
	logger.Init()
}

func main() {
	port := env.GetStringOrDefault("PORT", "8080")

	dbAuth := env.GetStringOrDefault("DB_URI", "")
	if dbAuth == "" {
		logger.Error("DB_URI environment variable is not set")
		os.Exit(1)
	}

	dbConn, err := db.Connect(dbAuth)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			logger.Error("Failed to close database connection", "error", err)
		}
	}()

	// DI: Querier → Repository → Handler
	q := dbq.New(dbConn)
	h := initHandlers(q, dbConn)

	r := routes.NewRouter(h)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Starting server", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exited gracefully")
}

func initHandlers(q dbq.Querier, conn *sql.DB) routes.Handlers {
	taskRepo := task_repository.NewTaskRepository(q)
	orgRepo := organization_repository.NewOrganizationRepository(q)
	projectRepo := project_repository.NewProjectRepository(q)
	promptRepo := prompt_repository.NewPromptRepository(q)
	versionRepo := prompt_repository.NewVersionRepository(q)

	return routes.Handlers{
		Health:       healthHandler(conn),
		Task:         tasks.NewTaskHandler(taskRepo),
		Organization: organizations.NewOrganizationHandler(orgRepo),
		Project:      projects.NewProjectHandler(projectRepo),
		Prompt:       prompts.NewPromptHandler(promptRepo, versionRepo),
	}
}

func healthHandler(conn *sql.DB) http.HandlerFunc {
	type healthResponse struct {
		Status   string `json:"status"`
		Database string `json:"database"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		dbStatus := "ok"
		if err := conn.PingContext(ctx); err != nil {
			logger.Error("health check failed", "error", err)
			dbStatus = "unhealthy"
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(healthResponse{
				Status:   "unhealthy",
				Database: dbStatus,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(healthResponse{
			Status:   "ok",
			Database: dbStatus,
		})
	}
}
