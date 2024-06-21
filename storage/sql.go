package storage

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"todo-service/types"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type SQLAdapter struct {
	DB *gorm.DB
}

var sqlAdapterLock = &sync.Mutex{}
var sqlAdapterInstance *SQLAdapter

func GetSQLAdapterInstance() *SQLAdapter {
	if sqlAdapterInstance == nil {
		sqlAdapterLock.Lock()
		defer sqlAdapterLock.Unlock()
		if sqlAdapterInstance == nil {
			sqlAdapterInstance = &SQLAdapter{}
			sqlAdapterInstance.OpenConnection()
		}
	}
	return sqlAdapterInstance
}

func (s *SQLAdapter) OpenConnection() {
	var err error
	provider := viper.GetString("storage.provider")
	config := viper.GetStringMapString("storage.config")

	gormConf := gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   fmt.Sprintf("%s.", viper.GetString("storage.config.schema")),
			SingularTable: false,
		},
		Logger: logger.Default.LogMode(logger.Silent),
	}

	switch provider {
	case "postgresql":
		dsn := new(bytes.Buffer)

		for key, value := range config {
			if key != "schema" {
				fmt.Fprintf(dsn, "%s=%s ", key, value)
			}
		}
		s.DB, err = gorm.Open(postgres.New(postgres.Config{DSN: dsn.String(), PreferSimpleProtocol: true}), &gormConf)
	case "mysql":
		dsn := new(bytes.Buffer)
		fmt.Fprintf(dsn, "%s:%s@tcp(%s:%s)/%s", config["user"], config["password"], config["host"], config["port"], config["dbname"])
		s.DB, err = gorm.Open(mysql.New(mysql.Config{DSN: dsn.String()}), &gormConf)
	case "sqlite":
		s.DB, err = gorm.Open(sqlite.Open(config["path"]), &gormConf)
	default:
		log.Fatal("this SQL provider is not supported, supported providers are: postgresql, mysql, and sqlite")
	}

	if err != nil {
		log.Fatalf("failed to open a database connnection: %s", err.Error())
	}
}

func (s *SQLAdapter) Execute(statement string) error {
	result := s.DB.Exec(statement)
	if result.Error != nil {
		return fmt.Errorf("failed to execute statement %s: %v", statement, result.Error)
	}
	return nil
}

func (s *SQLAdapter) ListTodos() ([]types.Todo, error) {
	todos := []types.Todo{}
	result := s.DB.Find(&todos)
	return todos, result.Error
}

func (s *SQLAdapter) GetTodo(id string) (types.Todo, error) {
	todo := types.Todo{}
	result := s.DB.Where("Id = ?", id).Find(&todo)
	if result.RowsAffected == 0 {
		return todo, ErrNotFound
	}
	return todo, result.Error
}

func (s *SQLAdapter) DeleteTodo(id string) error {
	result := s.DB.Where("Id = ?", id).Delete(&types.Todo{})
	return result.Error
}

func (s *SQLAdapter) CreateTodo(todo types.Todo) error {
	result := s.DB.Create(&todo)
	return result.Error
}
