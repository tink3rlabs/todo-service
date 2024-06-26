package storage

import (
	"fmt"
	"log"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type MigrationFile struct {
	Description string
	Migrations  []Migration
}

type Migration struct {
	Migrate  string
	Rollback string
}

type DatabaseMigration struct {
	storageType     string
	storageProvider string
	storage         StorageAdapter
}

func NewDatabaseMigration() *DatabaseMigration {
	s := StorageAdapterFactory{}
	storageAdapter, err := s.GetInstance(DEFAULT)
	if err != nil {
		log.Fatalf("failed to create DatabaseMigration instance: %s", err.Error())
		return nil
	}
	m := DatabaseMigration{
		storage:         storageAdapter,
		storageType:     viper.GetString("storage.type"),
		storageProvider: viper.GetString("storage.provider"),
	}
	return &m
}

func (m *DatabaseMigration) getMigrationFiles() (map[string]MigrationFile, error) {
	var err error
	migrations := map[string]MigrationFile{}
	path := fmt.Sprintf("config/migrations/%s", m.storageProvider)
	files, _ := ConfigFs.ReadDir(path)

	for _, f := range files {
		var contents []byte
		contents, err = ConfigFs.ReadFile(filepath.Join(path, f.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %v", f.Name(), err)
		}
		mf := MigrationFile{}
		err = yaml.Unmarshal([]byte(contents), &mf)
		if err != nil {
			return nil, fmt.Errorf("failed to parse migration file %s: %v", f.Name(), err)
		}
		migrations[f.Name()] = mf
	}
	return migrations, nil
}

func (m *DatabaseMigration) createSchema() error {
	if m.storageProvider != "sqlite" {
		statement := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", viper.GetString("storage.config.schema"))
		return m.storage.Execute(statement)
	}
	return nil
}

func (m *DatabaseMigration) createMigrationTable() error {
	var statement string
	switch m.storageProvider {
	case "postgresql":
		statement = "CREATE TABLE IF NOT EXISTS migrations (id NUMERIC PRIMARY KEY, name TEXT, description TEXT, timestamp NUMERIC)"
	case "mysql":
		statement = "CREATE TABLE IF NOT EXISTS migrations (id INT PRIMARY KEY, name TEXT, description TEXT, timestamp BIGINT)"
	case "sqlite":
		statement = "CREATE TABLE IF NOT EXISTS migrations (id INTEGER PRIMARY KEY, name TEXT, description TEXT, timestamp INTEGER)"
	}
	return m.storage.Execute(statement)
}

func (m *DatabaseMigration) updateMigrationTable(id int, name string, desc string) error {
	statement := fmt.Sprintf(`INSERT INTO migrations VALUES(%v, '%v', '%v', %v)`, id, name, desc, time.Now().UnixMilli())
	return m.storage.Execute(statement)
}

func (m *DatabaseMigration) getLatestMigration() (int, error) {
	var statement string
	var latestMigration int
	switch m.storageType {
	case "sql":
		statement = "SELECT max(id) from migrations"
		a := GetSQLAdapterInstance()
		result := a.DB.Raw(statement).Scan(&latestMigration)
		if result.Error != nil {
			//either a real issue or there are no migrations yet check if we can query the migration table
			var count int
			statement = "SELECT count(*) from migrations"
			a := GetSQLAdapterInstance()
			countResult := a.DB.Raw(statement).Scan(&count)
			if countResult.Error != nil {
				return latestMigration, result.Error
			}
		}
	}
	return latestMigration, nil
}

func (m *DatabaseMigration) rollbackMigration(migration MigrationFile) error {
	var err error
	slices.Reverse(migration.Migrations)
	for _, s := range migration.Migrations {
		err = m.storage.Execute(s.Rollback)
		if err != nil {
			break
		}
	}
	return err
}

func (m *DatabaseMigration) runMigrations(migrations map[string]MigrationFile) {
	log.Println("Getting last migration applied")
	rollback := false
	latestMigrationId, err := m.getLatestMigration()
	if err != nil {
		log.Fatalf("failed to get latest migration: %v", err)
	}

	//iterating over a map is randomized so we need to make sure we use the correct order of migrations
	keys := make([]string, 0, len(migrations))
	for k := range migrations {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		migrationId, err := strconv.Atoi(strings.Split(k, "__")[0])
		if err != nil {
			log.Fatalf("failed to determine migration id: %v", err)
		}
		if migrationId > latestMigrationId {
			mf := migrations[k]
			for _, stmt := range mf.Migrations {
				err := m.storage.Execute(stmt.Migrate)
				if err != nil {
					log.Printf("failed to execute migration statement: %v", err)
					log.Printf("attempting to rollback migration: %s", k)
					rollback = true
					err = m.rollbackMigration(mf)
					if err != nil {
						log.Fatalf("failed to rollback migration: %v", err)
					}
					log.Print("rollback successful")
					break
				}
			}
			if rollback {
				break
			}
			log.Printf("updating migration table for %v", k)
			err = m.updateMigrationTable(migrationId, k, mf.Description)
			if err != nil {
				log.Fatalf("failed to update migration table: %v", err)
			}
		}
	}
}

func (m *DatabaseMigration) Migrate() {
	if m.storageType == "memory" {
		log.Println("using memory storage adapter, migrations are not needed")
	} else {
		log.Println("using a persistent storage adapter, executing migrations")
		migrations, err := m.getMigrationFiles()
		if err != nil {
			log.Fatalf("failed to get migration files: %v", err)
		}
		log.Println("creating schema")
		err = m.createSchema()
		if err != nil {
			log.Fatalf("failed to create schema: %v", err)
		}
		log.Println("creating migration table")
		err = m.createMigrationTable()
		if err != nil {
			log.Fatalf("failed to create migration table: %v", err)
		}
		m.runMigrations(migrations)
		log.Println("finished running migrations")
	}
}
