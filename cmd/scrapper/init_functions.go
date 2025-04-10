package main

import (
	"LinkTracker/internal/application/scrapper"
	"LinkTracker/internal/infrastructure/repository/postgresql/goqurepo"
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

func InitRepositories(ctx context.Context, dbConfig application.DBConfig, accessType string) (
	scrapper.UserRepo, scrapper.LinkRepo, scrapper.StateRepo, error) {
	connStr := "postgres://" + dbConfig.PostgresUser +
		":" + dbConfig.PostgresPassword +
		"@localhost:5432/" + dbConfig.PostgresDB + "?pool_max_conns=10"

	pool, err := pgxrepo.NewPool(ctx, connStr)
	if err != nil {
		fmt.Printf("Error creating pool: %v\n", err)
		return nil, nil, nil, err
	}

	var (
		userRepo  scrapper.UserRepo
		linkRepo  scrapper.LinkRepo
		stateRepo scrapper.StateRepo
	)
	if accessType == "GOQU" {
		userRepo = goqurepo.NewUserRepoGoqu(pool)
		linkRepo = goqurepo.NewLinkRepoGoqu(pool)
		stateRepo = goqurepo.NewStateRepoGoqu(pool)

		return userRepo, linkRepo, stateRepo, nil
	}

	userRepo = pgxrepo.NewUserRepo(pool)
	linkRepo = pgxrepo.NewLinkRepo(pool)
	stateRepo = pgxrepo.NewStateRepoPgx(pool)

	return userRepo, linkRepo, stateRepo, nil
}
