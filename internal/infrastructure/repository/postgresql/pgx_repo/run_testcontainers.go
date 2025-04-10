package pgxrepo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func getMigrationsPath() (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("не удалось получить рабочий каталог: %w", err)
	}

	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(workingDir)))))
	migrationsPath := filepath.Join(projectRoot, "migrations")

	return migrationsPath, nil
}

func createNetwork(ctx context.Context) (tc.Network, string, error) { //nolint: staticcheck, gocritic // half of the library is deprecated
	netw, err := network.New(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("не удалось создать сеть: %w", err)
	}

	return netw, netw.Name, nil
}

func startPostgresContainer(ctx context.Context, networkName string) (tc.Container, string, error) {
	postgresReq := tc.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
		Networks:   []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"postgres"},
		},
	}

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	if err != nil {
		return nil, "", fmt.Errorf("не удалось запустить контейнер postgres: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("не удалось получить хост postgres: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, "", fmt.Errorf("не удалось получить порт postgres: %w", err)
	}

	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, mappedPort.Port())

	return container, dsn, nil
}

func createPgxPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать пул подключений: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("не удалось пинговать postgres: %w", err)
	}

	return pool, nil
}

func runMigrations(ctx context.Context, networkName string) error {
	migrationsPath, err := getMigrationsPath()
	if err != nil {
		return fmt.Errorf("не удалось вычислить путь к миграциям: %w", err)
	}

	dsn := "postgres://testuser:testpass@postgres:5432/testdb?sslmode=disable"
	migrateReq := tc.ContainerRequest{
		Image: "migrate/migrate:v4.15.2",
		Cmd: []string{
			"--path", "/migrations",
			"--database", dsn,
			"up",
		},
		Networks: []string{networkName},
		Mounts: tc.Mounts(
			tc.BindMount(migrationsPath, "/migrations"), //nolint: staticcheck,gocritic // half of the library is deprecated
		),
		WaitingFor: wait.ForExit(),
	}

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: migrateReq,
		Started:          true,
	})
	if err != nil {
		return fmt.Errorf("не удалось запустить контейнер миграций: %w", err)
	}

	return container.Terminate(ctx)
}

func RunTestContainers(ctx context.Context) (*pgxpool.Pool, func(), error) {
	netw, networkName, err := createNetwork(ctx)
	if err != nil {
		return nil, nil, err
	}

	pgContainer, dsn, err := startPostgresContainer(ctx, networkName)
	if err != nil {
		_ = netw.Remove(ctx)
		return nil, nil, err
	}

	pool, err := createPgxPool(ctx, dsn)
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)

		return nil, nil, err
	}

	if err := runMigrations(ctx, networkName); err != nil {
		pool.Close()

		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)

		return nil, nil, err
	}

	cleanup := func() {
		pool.Close()

		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)
	}

	return pool, cleanup, nil
}
