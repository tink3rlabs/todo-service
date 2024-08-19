package storage

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log/slog"
	"sync"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"todo-service/types"
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
	case string(POSTGRESQL):
		dsn := new(bytes.Buffer)

		for key, value := range config {
			if key != "schema" {
				fmt.Fprintf(dsn, "%s=%s ", key, value)
			}
		}
		s.DB, err = gorm.Open(postgres.New(postgres.Config{DSN: dsn.String(), PreferSimpleProtocol: true}), &gormConf)
	case string(MYSQL):
		dsn := new(bytes.Buffer)
		fmt.Fprintf(dsn, "%s:%s@tcp(%s:%s)/%s", config["user"], config["password"], config["host"], config["port"], config["dbname"])
		s.DB, err = gorm.Open(mysql.New(mysql.Config{DSN: dsn.String()}), &gormConf)
	case string(SQLITE):
		s.DB, err = gorm.Open(sqlite.Open(config["path"]), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	default:
		slog.Error("this SQL provider is not supported, supported providers are: postgresql, mysql, and sqlite")
	}

	if err != nil {
		slog.Error("failed to open a database connection", slog.Any("error", err.Error()))
	}
}

func (s *SQLAdapter) Execute(statement string) error {
	result := s.DB.Exec(statement)
	if result.Error != nil {
		return fmt.Errorf("failed to execute statement %s: %v", statement, result.Error)
	}
	return nil
}

func (s *SQLAdapter) Ping() error {
	db, err := s.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	return db.Ping()
}

func (s *SQLAdapter) ListTodos(limit int, cursor string) ([]types.Todo, string, error) {
	todos := []types.Todo{}
	nextId := ""

	id, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return todos, "", fmt.Errorf("failed to decode next cursor: %v", err)
	}

	// Get one extra item to be able to set that item's Id as the cursor for the next request
	result := s.DB.Limit(limit+1).Where("id >= ?", string(id)).Find(&todos)

	// If we have a full list, set the Id of the extra last item as the next cursor and remove it from the list of items to return
	if len(todos) == limit+1 {
		nextId = base64.StdEncoding.EncodeToString([]byte(todos[len(todos)-1].Id))
		todos = todos[:len(todos)-1]
	}

	return todos, nextId, result.Error
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
