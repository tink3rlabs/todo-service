package storage

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3"
	"github.com/scylladb/gocqlx/v3/table"
	"github.com/spf13/viper"

	"todo-service/internal/logger"
	"todo-service/types"
)

type CassandraAdapter struct {
	DB *gocqlx.Session
}

var cassandraAdapterLock = &sync.Mutex{}
var cassandraAdapterInstance *CassandraAdapter

var todoTable = table.New(table.Metadata{
	Name:    "todos",
	Columns: []string{"id", "summary"},
	PartKey: []string{"id"},
})

func GetCassandraAdapterInstance() *CassandraAdapter {
	if cassandraAdapterInstance == nil {
		cassandraAdapterLock.Lock()
		defer cassandraAdapterLock.Unlock()
		if cassandraAdapterInstance == nil {
			cassandraAdapterInstance = &CassandraAdapter{}
			cassandraAdapterInstance.OpenConnection()
		}
	}
	return cassandraAdapterInstance
}

func (s *CassandraAdapter) OpenConnection() {
	var err error
	config := viper.GetStringMap("storage.config")

	inputHosts := config["hosts"].([]interface{})
	hosts := make([]string, len(inputHosts))
	for i, v := range inputHosts {
		hosts[i] = fmt.Sprintf(`%v:%v`, v, config["port"])
	}

	// Create gocql cluster.
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = config["keyspace"].(string)
	// Wrap session on creation, gocqlx session embeds gocql.Session pointer.
	session, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		logger.Fatal("failed to open a database connection", slog.Any("error", err.Error()))
	}
	s.DB = &session
}

func (s *CassandraAdapter) Execute(statement string) error {
	err := s.DB.ExecStmt(statement)
	if err != nil {
		return fmt.Errorf("failed to execute statement %s: %v", statement, err)
	}
	return nil
}

func (s *CassandraAdapter) Ping() error {
	statement := "SELECT UUID() FROM SYSTEM.LOCAL"
	return s.Execute(statement)
}

func (s *CassandraAdapter) ListTodos(limit int, cursor string) ([]types.Todo, string, error) {
	todos := []types.Todo{}

	pageState, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return todos, "", fmt.Errorf("failed to decode next cursor: %v", err)
	}

	q := s.DB.Query(todoTable.SelectAll())
	defer q.Release()

	q.PageState(pageState)
	q.PageSize(limit)
	iter := q.Iter()

	err = iter.Select(&todos)

	return todos, base64.StdEncoding.EncodeToString(iter.PageState()), err
}

func (s *CassandraAdapter) GetTodo(id string) (types.Todo, error) {
	todo := types.Todo{}
	q := s.DB.Query(todoTable.Get()).Bind(id)
	fmt.Println(q)
	err := q.GetRelease(&todo)
	if (err != nil) && (err.Error() == "not found") {
		return todo, ErrNotFound
	}
	return todo, err
}

func (s *CassandraAdapter) DeleteTodo(id string) error {
	q := s.DB.Query(todoTable.Delete()).Bind(id)
	return q.ExecRelease()
}

func (s *CassandraAdapter) CreateTodo(todo types.Todo) error {
	q := s.DB.Query(todoTable.Insert()).BindStruct(todo)
	return q.ExecRelease()
}
