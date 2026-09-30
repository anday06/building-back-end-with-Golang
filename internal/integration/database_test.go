//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"task-management-api/internal/database"
)

func startPostgresContainer(ctx context.Context) (container testcontainers.Container, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("could not start Testcontainers provider: %v", recovered)
		}
	}()
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:16-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_DB":       "tasks",
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "postgres",
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections"),
		},
		Started: true,
	})
}

func TestPostgresMigrationsWithTestcontainers(t *testing.T) {
	ctx := context.Background()
	container, err := startPostgresContainer(ctx)
	if err != nil {
		t.Skipf("Docker/Testcontainers is unavailable: %v", err)
	}
	defer container.Terminate(ctx)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	dsn := fmt.Sprintf("host=%s user=postgres password=postgres dbname=tasks port=%s sslmode=disable", host, port.Port())

	db, err := database.Connect(dsn)
	require.NoError(t, err)
	require.NoError(t, database.Migrate(db))
}
