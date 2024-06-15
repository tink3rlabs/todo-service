package storage

import (
	"sync"
	"todo-service/types"

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
	var err error
	//Create a new Postgresql database connection
	dsn := "host=host.docker.internal user=gorm password=gorm dbname=gorm port=5432"
	// Open a connection to the database
	s.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
}

func (s *SQLAdapter) ListTodos() []types.Todo {
	return []types.Todo{}
}

func (s *SQLAdapter) GetTodo(id string) (types.Todo, error) {
	return types.Todo{}, ErrNotFound
}

func (s *SQLAdapter) DeleteTodo(id string) {}

func (s *SQLAdapter) CreateTodo(todo types.Todo) {}
