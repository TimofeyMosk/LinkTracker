package main

import (
	"context"
	"fmt"

	"LinkTracker/internal/application"
	"LinkTracker/internal/application/scrapper/linkchecker"
	"LinkTracker/internal/infrastructure/clients"
	pgxrepo "LinkTracker/internal/infrastructure/repository/postgresql/pgx_repo"
)

func InitLinksSourceHandlers() []linkchecker.LinkSourceHandler {
	return []linkchecker.LinkSourceHandler{
		clients.NewGitHubHTTPClient(),
		clients.NewStackOverflowHTTPClient(),
	}
}

func InitRepositories(ctx context.Context, dbConfig application.DBConfig) (*pgxrepo.UserRepoPgx, *pgxrepo.LinkRepoPgx, *pgxrepo.StateRepoPgx, error) {
	connStr := "postgres://" + dbConfig.PostgresUser +
		":" + dbConfig.PostgresPassword +
		"@localhost:5432/" + dbConfig.PostgresDB + "?pool_max_conns=10"

	pool, err := pgxrepo.NewPool(ctx, connStr)
	if err != nil {
		fmt.Printf("Error creating pool: %v\n", err)
		return nil, nil, nil, err
	}

	userRepo := pgxrepo.NewUserRepo(pool)
	linkRepo := pgxrepo.NewLinkRepo(pool)
	stateManager := pgxrepo.NewStateRepoPgx(pool)

	return userRepo, linkRepo, stateManager, nil
}
