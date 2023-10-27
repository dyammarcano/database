//go:build windows

package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/ory/dockertest/v3"
	"log"
	"testing"
	"time"
)

var (
	db     *sql.DB
	config = &Config{
		DriverName: "mysql",
		User:       "test",
		Pass:       "test",
		Host:       "localhost",
		Name:       "oracleDB",
		Port:       3306, // MySQL default port
	}
)

func setup() (*dockertest.Resource, *dockertest.Pool, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, fmt.Errorf("could not construct pool: %w", err)
	}

	pool.MaxWait = time.Minute * 2

	err = pool.Client.Ping()
	if err != nil {
		return nil, nil, fmt.Errorf("could not connect to Docker: %w", err)
	}

	resource, err := pool.Run("mysql", "5.7", []string{"MYSQL_ROOT_PASSWORD=secret"})
	if err != nil {
		return nil, nil, fmt.Errorf("could not start resource: %w", err)
	}

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.User, config.Pass, config.Host, config.Port, config.Name)
	db, err = sql.Open(config.DriverName, connStr)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open database connection: %w", err)
	}

	if err := pool.Retry(func() error {
		return db.Ping()
	}); err != nil {
		return nil, nil, fmt.Errorf("could not connect to database: %w", err)
	}

	return resource, pool, nil
}

func teardown(pool *dockertest.Pool, resource *dockertest.Resource) {
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}
}

func TestNewDatabase_Connect(t *testing.T) {
	resource, pool, err := setup()
	if err != nil {
		t.Fatal(err)
	}
	defer teardown(pool, resource)

	err = db.Ping()
	if err != nil {
		t.Errorf("Failed to ping database: %v", err)
	}
}
