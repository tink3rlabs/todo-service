package storage

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"todo-service/types"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	provider := viper.GetString("storage.provider")

	switch provider {
	case "postgresql":
		var err error
		config := viper.GetStringMapString("storage.config")
		dsn := new(bytes.Buffer)

		for key, value := range config {
			fmt.Fprintf(dsn, "%s=%s ", key, value)
		}

		s.DB, err = gorm.Open(postgres.New(postgres.Config{
			DSN:                  dsn.String(),
			PreferSimpleProtocol: true}), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to open a database connnection: %s", err.Error())
		}
	default:
		log.Fatal("this SQL provider is not supported, supported providers are: postgresql and mysql")
	}
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
