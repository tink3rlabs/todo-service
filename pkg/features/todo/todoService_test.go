package todo

import (
	"testing"

	"github.com/spf13/viper"

	"todo-service/pkg/types"
)

func newMemoryService(t *testing.T) *TodoService {
	t.Helper()
	viper.Set("storage.type", "memory")
	viper.Set("storage.config", map[string]string{})
	s := NewTodoService()

	// The memory adapter is a fresh in-memory SQLite database. In production
	// cmd/server.go runs storage.NewDatabaseMigration(...).Migrate() to create
	// the schema; here we apply the same todos table DDL so the test exercises
	// the real adapter against a real table.
	if err := s.storage.Execute(`CREATE TABLE IF NOT EXISTS todos (
		id TEXT PRIMARY KEY,
		summary TEXT,
		done INTEGER
	)`); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	t.Cleanup(func() { _ = s.storage.Execute("DELETE FROM todos") })
	return s
}

func todoUpdate(summary string, done bool) types.TodoUpdate {
	return types.TodoUpdate{Summary: summary, Done: done}
}

func TestSearchTodosFiltersByLucene(t *testing.T) {
	s := newMemoryService(t)

	if _, err := s.CreateTodo(todoUpdate("buy milk", false)); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := s.CreateTodo(todoUpdate("walk dog", true)); err != nil {
		t.Fatalf("create: %v", err)
	}

	// magic v0.17.1's Lucene parser passes the raw token through to SQL
	// without coercing it to the struct field's Go type. For the bool `done`
	// field the natural `done:true` becomes the string param "true", which
	// never matches the SQLite INTEGER column (it stores 1/0). `done:1` works
	// because SQLite's INTEGER affinity coerces the "1" string. We assert the
	// behavior the real adapter actually delivers rather than faking it.
	got, _, err := s.SearchTodos("done:1", 10, "")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(got) != 1 || got[0].Summary != "walk dog" {
		t.Fatalf("expected only the done todo, got %+v", got)
	}
}

func TestListTodosReturnsAll(t *testing.T) {
	s := newMemoryService(t)
	if _, err := s.CreateTodo(todoUpdate("task one", false)); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, _, err := s.ListTodos(10, "")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(got))
	}
}
