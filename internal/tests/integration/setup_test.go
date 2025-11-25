package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgresContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"avito/internal/handlers"
	"avito/internal/repository"
	"avito/internal/service"
)

type TestEnvironment struct {
	DB           *sql.DB
	Router       http.Handler
	Server       *httptest.Server
	Container    *postgresContainer.PostgresContainer
	TeamHandler  *handlers.TeamHandler
	UserHandler  *handlers.UserHandler
	PRHandler    *handlers.PullRequestHandler
	StatsHandler *handlers.StatisticsHandler
	TeamService  *service.TeamService
	UserService  *service.UserService
	PRService    *service.PullRequestService
	StatsService *service.StatisticsService
	TeamRepo     repository.TeamRepository
	UserRepo     repository.UserRepository
	PRRepo       repository.PullRequestRepository
	StatsRepo    repository.StatisticsRepository
	TxMgr        repository.TransactionManager
}

func setupTestDB(t *testing.T) (*sql.DB, *postgresContainer.PostgresContainer, func()) {
	ctx := context.Background()

	postgresC, err := postgresContainer.Run(ctx, "postgres:15-alpine", postgresContainer.WithDatabase("test_db"), postgresContainer.WithUsername("test"), postgresContainer.WithPassword("test"), testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(60*time.Second)))
	require.NoError(t, err)

	connStr, err := postgresC.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	runMigrations(t, db)

	cleanup := func() {
		if db != nil {
			db.Close()
		}
		if postgresC != nil {
			if err := postgresC.Terminate(ctx); err != nil {
				t.Logf("Failed to terminate container: %v", err)
			}
		}
	}

	return db, postgresC, cleanup
}

func runMigrations(t *testing.T, db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	require.NoError(t, err)

	m, err := migrate.NewWithDatabaseInstance("file://../../../migrations", "postgres", driver)
	require.NoError(t, err)

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}
}

func setupTestEnvironment(t *testing.T) *TestEnvironment {
	db, container, cleanup := setupTestDB(t)

	t.Cleanup(cleanup)

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

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	return &TestEnvironment{DB: db, Router: router, Server: server, Container: container, TeamHandler: teamHandler, UserHandler: userHandler, PRHandler: prHandler, StatsHandler: statsHandler, TeamService: teamService, UserService: userService, PRService: prService, StatsService: statsService, TeamRepo: teamRepo, UserRepo: userRepo, PRRepo: prRepo, StatsRepo: statsRepo, TxMgr: txMgr}
}

func cleanDatabase(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`TRUNCATE TABLE pr_reviewers CASCADE; TRUNCATE TABLE pull_requests CASCADE; TRUNCATE TABLE users CASCADE; TRUNCATE TABLE teams CASCADE;`)
	require.NoError(t, err)
}

func (env *TestEnvironment) BaseURL() string {
	return env.Server.URL
}

func (env *TestEnvironment) URL(path string) string {
	return fmt.Sprintf("%s%s", env.Server.URL, path)
}
