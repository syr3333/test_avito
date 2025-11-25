package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"avito/internal/handlers"
	"avito/internal/logging"
	"avito/internal/repository"
	"avito/internal/service"
	"avito/pkg/config"
)

func main() {
	logFilePath := os.Getenv("LOG_FILE_PATH")

	logging.SetUpLogger(logFilePath)

	cfg, err := config.Load()
	if err != nil {
		logging.Error("Failed to load config:", err)
		os.Exit(1)
	}
	logging.Info("Config loaded, port:", cfg.ServerPort)

	db, err := sql.Open("postgres", cfg.DBConnectionString())
	if err != nil {
		logging.Error("Failed to connect to database:", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logging.Error("Failed to ping database:", err)
		os.Exit(1)
	}
	logging.Info("Database connection established")

	teamRepo := repository.NewTeamRepository(db)
	userRepo := repository.NewUserRepository(db)
	prRepo := repository.NewPullRequestRepository(db)
	statsRepo := repository.NewStatisticsRepository(db)
	txMgr := repository.NewTransactionManager(db)

	teamService := service.NewTeamService(teamRepo, userRepo, txMgr)
	userService := service.NewUserService(userRepo)
	prService := service.NewPullRequestService(prRepo, userRepo, txMgr)
	statsService := service.NewStatisticsService(statsRepo)

	teamHandler := handlers.NewTeamHandler(teamService)
	userHandler := handlers.NewUserHandler(userService, prRepo)
	prHandler := handlers.NewPullRequestHandler(prService)
	statsHandler := handlers.NewStatisticsHandler(statsService)

	router := handlers.Router(teamHandler, userHandler, prHandler, statsHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logging.Info("Server starting on port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Error("Server failed to start:", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logging.Info("Server shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logging.Error("Server forced to shutdown:", err)
	}

	logging.Info("Server exited")
}
