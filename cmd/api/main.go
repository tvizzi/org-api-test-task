package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tvizzi/org-api-test-task/internal/handler"
	"github.com/tvizzi/org-api-test-task/internal/middleware"
	"github.com/tvizzi/org-api-test-task/internal/repository"
	"github.com/tvizzi/org-api-test-task/internal/service"

	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		slog.Error("DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	sqlDB, _ := db.DB()
	if err := runMigrations(sqlDB); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations applied successfully")

	txManager := repository.NewTransactionManager(db)
	deptRepo := repository.NewDepartmentRepository(db)
	empRepo := repository.NewEmployeeRepository(db)

	deptService := service.NewDepartmentService(deptRepo, empRepo, txManager)
	empService := service.NewEmployeeService(empRepo, deptRepo)

	deptHandler := handler.NewDepartmentHandler(deptService)
	empHandler := handler.NewEmployeeHandler(empService)

	mux := http.NewServeMux()
	deptHandler.RegisterRoutes(mux)
	empHandler.RegisterRoutes(mux)

	loggedMux := middleware.Logging(logger)(mux)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           loggedMux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		slog.Info("server listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
	} else {
		slog.Info("server stopped gracefully")
	}
}

func runMigrations(db *sql.DB) error {
	goose.SetDialect("postgres")
	return goose.Up(db, "migrations")
}
