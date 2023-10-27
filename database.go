package database

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/godror/godror"
	_ "github.com/godror/godror"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

type Database interface {
	Connect(config *Config) (*gorm.DB, error)
}

type Config struct {
	DriverName string
	FilePath   string
	User       string
	Pass       string
	Host       string
	Name       string
	Port       int
}

type database struct{}

func NewDatabase() Database {
	return &database{}
}

func (db *database) Connect(config *Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch config.DriverName {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.User, config.Pass, config.Host, config.Port, config.Name)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.Host, config.Port, config.User, config.Pass, config.Name)
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(config.FilePath)
	case "oracle":
		params := connectParamsFromConfig(config)
		dialector = postgres.New(postgres.Config{Conn: sql.OpenDB(params)})
	default:
		return nil, fmt.Errorf("invalid database driver: %s", config.DriverName)
	}

	dbInstance, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := dbInstance.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return dbInstance, nil
}

func connectParamsFromConfig(cfg *Config) driver.Connector {
	var params godror.ConnectionParams
	params.Username = cfg.User
	params.Password = godror.NewPassword(cfg.Pass)
	params.ConnectString = fmt.Sprintf("%s:%d/%s?connect_timeout=2", cfg.Host, cfg.Port, cfg.Name)
	params.SessionTimeout = 42 * time.Second
	params.Timezone = time.Local

	return godror.NewConnector(params)
}

//package database
//
//import (
//	"context"
//	"database/sql"
//	"fmt"
//	"github.com/godror/godror"
//	"time"
//)
//
//type (
//	ConnectionFunc func(config *Config) (*sql.DB, error)
//
//	Database interface {
//		Connect(config *Config) (*sql.DB, error)
//	}
//
//	Config struct {
//		DriverName string
//		FilePath   string
//		User       string
//		Pass       string
//		Host       string
//		Name       string
//		Port       int
//	}
//
//	database struct {
//		db  *sql.DB
//		ctx context.Context
//	}
//)
//
//var connections = map[string]ConnectionFunc{
//	"oracle":   connectOracle,
//	"Postgres": connectPostgres,
//	"mysql":    connectMySQL,
//	"sqlite":   connectSQLite,
//}
//
//func NewDatabase() Database {
//	return &database{}
//}
//
//func (db *database) Connect(config *Config) (*sql.DB, error) {
//	connFunc, ok := connections[config.DriverName]
//	if !ok {
//		return nil, fmt.Errorf("invalid database driver: %s", config.DriverName)
//	}
//	return connFunc(config)
//}
//
//func connectOracle(cfg *Config) (*sql.DB, error) {
//	var params godror.ConnectionParams
//	params.Username = cfg.User
//	params.Password = godror.NewPassword(cfg.Pass)
//	params.ConnectString = fmt.Sprintf("%s:%d/%s?connect_timeout=2", cfg.Host, cfg.Port, cfg.Name)
//	params.SessionTimeout = 42 * time.Second
//	params.Timezone = time.Local
//
//	return sql.OpenDB(godror.NewConnector(params)), nil
//}
//
//func connectPostgres(cfg *Config) (*sql.DB, error) {
//	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.Name)
//	return sql.Open("postgres", dsn)
//}
//
//func connectMySQL(cfg *Config) (*sql.DB, error) {
//	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.Name)
//	return sql.Open("mysql", dsn)
//}
//
//func connectSQLite(cfg *Config) (*sql.DB, error) {
//	return sql.Open("sqlite3", cfg.FilePath)
//}
