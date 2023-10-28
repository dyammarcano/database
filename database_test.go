//go:build windows

package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

var (
	db     *sql.DB
	config = &Config{
		DriverName: "mysql",
		User:       "root",
		Pass:       "secret",
		Host:       "localhost",
		Name:       "mysql",
	}
)

func setup() (*dockertest.Resource, *dockertest.Pool, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, fmt.Errorf("could not construct pool: %w", err)
	}

	pool.MaxWait = time.Minute * 2

	if err = pool.Client.Ping(); err != nil {
		return nil, nil, fmt.Errorf("could not connect to Docker: %w", err)
	}

	resource, err := pool.Run("mysql", "5.7", []string{"MYSQL_ROOT_PASSWORD=secret"})
	if err != nil {
		return nil, nil, fmt.Errorf("could not start resource: %w", err)
	}

	newDatabase := NewDatabase()
	config.Port = resource.GetPort("3306/tcp")

	if err := pool.Retry(func() error {
		var err error
		db, err = newDatabase.Connect(config)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		return nil, nil, fmt.Errorf("could not connect to Database: %w", err)
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
	assert.Nil(t, err)

	defer teardown(pool, resource)

	//create table
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `users` (`id` int(11) NOT NULL AUTO_INCREMENT,`name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,`email` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,`password` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,PRIMARY KEY (`id`)) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;")
	assert.Nil(t, err)

	//insert
	_, err = db.Exec("INSERT INTO `users` (`name`, `email`, `password`) VALUES ('John Doe', 'john-doe@email.com', 'supersecret')")
	assert.Nil(t, err)

	//select
	rows, err := db.Query("SELECT * FROM `users`")
	assert.Nil(t, err)

	for rows.Next() {
		var id int
		var name string
		var email string
		var password string
		err = rows.Scan(&id, &name, &email, &password)
		assert.Nil(t, err)
		assert.Equal(t, "John Doe", name)
	}

	//delete
	_, err = db.Exec("DELETE FROM `users` WHERE `name` = 'John Doe'")
	assert.Nil(t, err)

	//select
	rows, err = db.Query("SELECT * FROM `users`")
	assert.Nil(t, err)

	assert.True(t, !rows.Next())
}
