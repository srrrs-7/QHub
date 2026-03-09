package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api/src/infra/rds/consulting_repository"
	"api/src/infra/rds/executionlog_repository"
	"api/src/infra/rds/organization_repository"
	"api/src/infra/rds/project_repository"
	"api/src/infra/rds/prompt_repository"
	"api/src/infra/rds/tag_repository"
	"api/src/infra/rds/task_repository"
	"api/src/routes"
	"api/src/routes/admin"
	"api/src/routes/analytics"
	"api/src/routes/apikeys"
	"api/src/routes/consulting"
	"api/src/routes/evaluations"
	"api/src/routes/industries"
	"api/src/routes/logs"
	"api/src/routes/members"
	"api/src/routes/organizations"
	"api/src/routes/projects"
	"api/src/routes/prompts"
	"api/src/routes/tags"
	"api/src/routes/tasks"
	"api/src/routes/users"
	"api/src/services/batchservice"
	"api/src/services/diffservice"
	"api/src/services/lintservice"
	"utils/cache"
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

	// Redis cache (optional: enabled when REDIS_URL is set)
	var cacheClient *cache.Client
	redisURL := env.GetStringOrDefault("REDIS_URL", "")
	if redisURL != "" {
		cacheClient = cache.New(redisURL)
		if cacheClient != nil && cacheClient.Available() {
			logger.Info("Redis cache enabled", "url", redisURL)
			defer cacheClient.Close()
		}
	} else {
		logger.Info("Redis cache disabled (set REDIS_URL to enable)")
	}

	// DI: Querier → Repository → Handler
	q := dbq.New(dbConn)
	h := initHandlers(q, dbConn, cacheClient)

	r := routes.NewRouter(h, q)

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

func initHandlers(q dbq.Querier, conn *sql.DB, cacheClient *cache.Client) routes.Handlers {
	taskRepo := task_repository.NewTaskRepository(q)
	orgRepo := organization_repository.NewOrganizationRepository(q)
	projectRepo := project_repository.NewProjectRepository(q)
	promptRepo := prompt_repository.NewPromptRepository(q)
	versionRepo := prompt_repository.NewVersionRepository(q)
	diffSvc := diffservice.NewDiffService(versionRepo, cacheClient)
	lintSvc := lintservice.NewLintService(versionRepo)
	logRepo := executionlog_repository.NewLogRepository(q)
	evalRepo := executionlog_repository.NewEvaluationRepository(q)
	sessionRepo := consulting_repository.NewSessionRepository(q)
	messageRepo := consulting_repository.NewMessageRepository(q)
	industryRepo := consulting_repository.NewIndustryConfigRepository(q)
	tagRepo := tag_repository.NewTagRepository(q)

	batchSvc := batchservice.NewBatchService(q)

	return routes.Handlers{
		Health:       healthHandler(conn, cacheClient),
		Task:         tasks.NewTaskHandler(taskRepo),
		Organization: organizations.NewOrganizationHandler(orgRepo),
		Project:      projects.NewProjectHandler(projectRepo),
		Prompt:       prompts.NewPromptHandler(promptRepo, versionRepo, diffSvc, lintSvc),
		Log:          logs.NewLogHandler(logRepo, evalRepo),
		Evaluation:   evaluations.NewEvaluationHandler(evalRepo),
		Consulting:   consulting.NewConsultingHandler(sessionRepo, messageRepo, industryRepo),
		Tag:          tags.NewTagHandler(tagRepo),
		Industry:     industries.NewIndustryHandler(industryRepo, q),
		Analytics:    analytics.NewAnalyticsHandler(q),
		ApiKey:       apikeys.NewApiKeyHandler(q),
		Member:       members.NewMemberHandler(q),
		User:         users.NewUserHandler(q),
		Admin:        admin.NewAdminHandler(batchSvc),
	}
}

func healthHandler(conn *sql.DB, cacheClient *cache.Client) http.HandlerFunc {
	type healthResponse struct {
		Status   string `json:"status"`
		Database string `json:"database"`
		Cache    string `json:"cache,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		resp := healthResponse{Status: "ok"}

		resp.Database = "ok"
		if err := conn.PingContext(ctx); err != nil {
			logger.Error("health check failed", "error", err)
			resp.Status = "unhealthy"
			resp.Database = "unhealthy"
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		if cacheClient != nil && cacheClient.Available() {
			if err := cacheClient.Ping(ctx); err != nil {
				resp.Cache = "unhealthy"
			} else {
				resp.Cache = "ok"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}
}
