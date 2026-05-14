package todo

import (
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/spf13/viper"

	"todo-service/pkg/types"

	"github.com/tink3rlabs/magic/logger"
	"github.com/tink3rlabs/magic/pubsub"
	"github.com/tink3rlabs/magic/storage"
	"github.com/tink3rlabs/magic/telemetry"
)

type TodoService struct {
	storage   storage.StorageAdapter
	created   telemetry.Counter
	publisher pubsub.Publisher
	topic     string
}

// WithPublisher attaches a pub/sub publisher; todo lifecycle events are published to topic.
func (t *TodoService) WithPublisher(p pubsub.Publisher, topic string) *TodoService {
	t.publisher = p
	t.topic = topic
	return t
}

// WithCreatedCounter attaches a metrics counter incremented on each successful create.
func (t *TodoService) WithCreatedCounter(c telemetry.Counter) *TodoService {
	t.created = c
	return t
}

func NewTodoService() *TodoService {
	storageAdapter, err := storage.StorageAdapterFactory{}.GetInstance(
		storage.StorageAdapterType(viper.GetString("storage.type")),
		viper.GetStringMapString("storage.config"),
	)

	if err != nil {
		logger.Fatal("failed to create TodoService instance", slog.Any("error", err.Error()))
	}
	t := TodoService{storage: storageAdapter}
	return &t
}

func (t *TodoService) ListTodos(limit int, cursor string) ([]types.Todo, string, error) {
	todos := []types.Todo{}
	next, err := t.storage.List(&todos, "id", map[string]any{}, limit, cursor)

	return todos, next, err
}

// SearchTodos returns todos matching a Lucene filter string, cursor-paginated.
// An empty filter returns everything (subject to limit/cursor).
func (t *TodoService) SearchTodos(filter string, limit int, cursor string) ([]types.Todo, string, error) {
	todos := []types.Todo{}
	next, err := t.storage.Search(&todos, "id", filter, limit, cursor)
	return todos, next, err
}

func (t *TodoService) GetTodo(id string) (types.Todo, error) {
	todo := types.Todo{}
	err := t.storage.Get(&todo, map[string]any{"id": id})
	return todo, err
}

func (t *TodoService) DeleteTodo(id string) error {
	return t.storage.Delete(&types.Todo{}, map[string]any{"id": id})
}

func (t *TodoService) UpdateTodo(todoToUpdate types.Todo) error {
	err := t.storage.Update(todoToUpdate, map[string]any{"id": todoToUpdate.Id})
	if err == nil {
		t.publishEvent("todo.updated", todoToUpdate)
	}
	return err
}

func (t *TodoService) CreateTodo(todoToCreate types.TodoUpdate) (types.Todo, error) {
	todo := types.Todo{}

	// Using UUIDv7 in order to easily support cursor based pagination without extra fields
	//
	// From the RFC (https://datatracker.ietf.org/doc/rfc9562/)
	//
	// UUIDv7 features a time-ordered value field derived from the widely
	// implemented and well-known Unix Epoch timestamp source, the number of
	// milliseconds since midnight 1 Jan 1970 UTC, leap seconds excluded.
	// Generally, UUIDv7 has improved entropy characteristics over UUIDv1
	// (Section 5.1) or UUIDv6 (Section 5.6).
	//
	// UUIDv7 values are created by allocating a Unix timestamp in
	// milliseconds in the most significant 48 bits and filling the
	// remaining 74 bits, excluding the required version and variant bits,
	// with random bits for each new UUIDv7 generated to provide uniqueness
	// as per Section 6.9.
	id, err := uuid.NewV7()
	if err != nil {
		return todo, err
	}

	todo.Id = id.String()
	todo.Summary = todoToCreate.Summary
	todo.Done = todoToCreate.Done

	err = t.storage.Create(todo)
	if err != nil {
		return todo, err
	}

	if t.created != nil {
		t.created.Add(1)
	}

	t.publishEvent("todo.created", todo)

	return todo, nil
}

func (t *TodoService) publishEvent(eventType string, todo types.Todo) {
	if t.publisher == nil {
		return
	}
	payload, err := json.Marshal(todo)
	if err != nil {
		slog.Error("failed to marshal todo event", slog.String("error", err.Error()))
		return
	}
	if err := t.publisher.Publish(t.topic, string(payload), map[string]any{"event_type": eventType}); err != nil {
		slog.Error("failed to publish todo event", slog.String("error", err.Error()), slog.String("event_type", eventType))
	}
}
