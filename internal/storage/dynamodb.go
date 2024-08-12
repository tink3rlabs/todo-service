package storage

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/spf13/viper"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"todo-service/internal/logger"
	"todo-service/types"
)

type DynamoDBAdapter struct {
	DB *dynamodb.Client
}

var dynamoDBAdapterLock = &sync.Mutex{}
var dynamoDBAdapterInstance *DynamoDBAdapter
var todosTable = "todos"

func GetDynamoDBAdapterInstance() *DynamoDBAdapter {
	if dynamoDBAdapterInstance == nil {
		dynamoDBAdapterLock.Lock()
		defer dynamoDBAdapterLock.Unlock()
		if dynamoDBAdapterInstance == nil {
			dynamoDBAdapterInstance = &DynamoDBAdapter{}
			dynamoDBAdapterInstance.OpenConnection()
		}
	}
	return dynamoDBAdapterInstance
}

func (s *DynamoDBAdapter) OpenConnection() {
	props := viper.GetStringMapString("storage.config")
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(props["region"]),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(props["access_key"], props["secret_key"], "")),
	)

	if err != nil {
		logger.Fatal("failed to open a database connection", slog.Any("error", err.Error()))
	}

	s.DB = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		if props["endpoint"] != "" {
			o.BaseEndpoint = aws.String(props["endpoint"])
		}
	})
}

func (s *DynamoDBAdapter) Execute(statement string) error {
	_, err := s.DB.ExecuteStatement(context.TODO(), &dynamodb.ExecuteStatementInput{Statement: &statement})
	if err != nil {
		return fmt.Errorf("failed to execute statement %s: %v", statement, err)
	}
	return nil
}

func (s *DynamoDBAdapter) Ping() error {
	// dynamodb is a managed service so as long as it responds to api calls we can consider it up
	_, err := s.DB.ListTables(context.TODO(), &dynamodb.ListTablesInput{})
	return err
}

func (s *DynamoDBAdapter) ListTodos(limit int, cursor string) ([]types.Todo, string, error) {
	todos := []types.Todo{}
	nextId := ""

	input := &dynamodb.ScanInput{
		TableName: aws.String(todosTable),
		Limit:     aws.Int32(int32(limit)),
	}

	id, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return todos, "", fmt.Errorf("failed to decode next cursor: %v", err)
	}

	if len(id) > 0 {
		m := map[string]string{}
		err = json.Unmarshal(id, &m)
		if err != nil {
			return todos, nextId, fmt.Errorf("failed to Unmarshal cursor, %v", err)
		}

		startKey, err := attributevalue.MarshalMap(m)
		if err != nil {
			return todos, nextId, fmt.Errorf("failed to marshal next cursor into dynamodb StartKey, %v", err)
		}

		input.ExclusiveStartKey = startKey
	}

	response, err := s.DB.Scan(context.TODO(), input)

	if err != nil {
		return todos, nextId, fmt.Errorf("failed to list todos, %v", err)
	}

	err = attributevalue.UnmarshalListOfMaps(response.Items, &todos)
	if err != nil {
		return todos, nextId, fmt.Errorf("failed to marshal scan response into todo list, %v", err)
	}

	if len(response.LastEvaluatedKey) != 0 {
		m := map[string]string{}
		err := attributevalue.UnmarshalMap(response.LastEvaluatedKey, &m)
		if err != nil {
			return todos, nextId, fmt.Errorf("failed to unmarshal LastEvaluatedKey, %v", err)
		}
		j, err := json.Marshal(m)
		if err != nil {
			return todos, nextId, fmt.Errorf("failed to encode LastEvaluatedKey into nextId cursor, %v", err)
		}
		nextId = base64.StdEncoding.EncodeToString([]byte(j))
	}
	return todos, nextId, err
}

func (s *DynamoDBAdapter) GetTodo(id string) (types.Todo, error) {
	todo := types.Todo{}
	key, err := attributevalue.MarshalMap(map[string]string{"id": id})
	if err != nil {
		return todo, fmt.Errorf("failed to marshal todo id into dynamodb attribute, %v", err)
	}

	response, err := s.DB.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(todosTable),
		Key:       key,
	})

	if err != nil {
		return todo, fmt.Errorf("failed to get todo, %v", err)
	}

	err = attributevalue.UnmarshalMap(response.Item, &todo)
	if err != nil {
		return todo, fmt.Errorf("failed to unmarshal dynamodb Get result into todo, %v", err)
	}

	if todo == (types.Todo{}) {
		return todo, ErrNotFound
	}

	return todo, nil
}

func (s *DynamoDBAdapter) DeleteTodo(id string) error {
	key, err := attributevalue.MarshalMap(map[string]string{"id": id})
	if err != nil {
		return fmt.Errorf("failed to marshal todo id into dynamodb attribute, %v", err)
	}

	_, err = s.DB.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(todosTable),
		Key:       key,
	})

	if err != nil {
		return fmt.Errorf("failed to delete todo, %v", err)
	}

	return nil
}

func (s *DynamoDBAdapter) CreateTodo(todo types.Todo) error {
	item, err := attributevalue.MarshalMapWithOptions(todo, func(eo *attributevalue.EncoderOptions) { eo.TagKey = "json" })
	if err != nil {
		return fmt.Errorf("failed to marshal todo into dynamodb item, %v", err)
	}

	_, err = s.DB.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(todosTable),
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to create todo: %v", err)
	}

	return nil
}
