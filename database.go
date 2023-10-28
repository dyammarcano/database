package database

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	//_ "github.com/IBM/go_ibm_db"
	//_ "github.com/denisenkom/go-mssqldb"
	"github.com/godror/godror"
	_ "github.com/godror/godror"
	//"gorm.io/driver/db2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"time"
)

type Interface interface {
	Connect(config *Config) (*sql.DB, error)
}

type Config struct {
	DriverName string
	FilePath   string
	User       string
	Pass       string
	Host       string
	Name       string
	Port       string
}

type Database struct{}

func NewDatabase() *Database {
	return &Database{}
}

func (db *Database) Connect(config *Config) (*sql.DB, error) {
	var dialector gorm.Dialector

	switch config.DriverName {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", config.User, config.Pass, config.Host, config.Port, config.Name)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", config.Host, config.Port, config.User, config.Pass, config.Name)
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(config.FilePath)
	case "oracle":
		params := connectParamsFromConfig(config)
		dialector = postgres.New(postgres.Config{Conn: sql.OpenDB(params)})
	case "sqlserver":
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?Database=%s", config.User, config.Pass, config.Host, config.Port, config.Name)
		dialector = sqlserver.Open(dsn)
	//case "db2":
	//	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.Host, config.Port, config.User, config.Pass, config.Name)
	//	dialector = db2.Open(dsn)
	default:
		return nil, fmt.Errorf("invalid Database driver: %s", config.DriverName)
	}

	dbInstance, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := dbInstance.DB()
	if err != nil {
		return nil, err
	}

	if config.DriverName != "sqlite" {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	return sqlDB, nil
}

func connectParamsFromConfig(cfg *Config) driver.Connector {
	var params godror.ConnectionParams
	params.Username = cfg.User
	params.Password = godror.NewPassword(cfg.Pass)
	params.ConnectString = fmt.Sprintf("%s:%s/%s?connect_timeout=2", cfg.Host, cfg.Port, cfg.Name)
	params.SessionTimeout = 42 * time.Second
	params.Timezone = time.Local

	return godror.NewConnector(params)
}
