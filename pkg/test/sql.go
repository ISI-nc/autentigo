package test

import (
	"database/sql"
	"fmt"
	"github.com/ory/dockertest/v3"
	"log"
	"os"
	"testing"
)

var (
	TestPool     *dockertest.Pool
	TestPostgresRsource *dockertest.Resource
	TestSql2KafkaResource *dockertest.Resource
)

func PostgresHost() string {
	if os.Getenv("CI") != "" {
		return "docker"
	}

	return "localhost"
}

func dockerHost() string {
	if os.Getenv("CI") != "" {
		return "docker:2375"
	}

	return ""
}

func StartPool(t *testing.T) {
	t.Logf("Starting docker pool...")

	var err error
	TestPool, err = dockertest.NewPool(dockerHost())
	if err != nil {
		log.Fatalf("Could not connect to docker daemon: %s", err)
	}
}

func StartPostgres(t *testing.T) *sql.DB {
	t.Logf("Starting postgres docker...")
	var pgdb *sql.DB
	var err error

	TestPostgresRsource, err = TestPool.Run("postgres", "alpine", []string{"POSTGRES_PASSWORD=postgres", "POSTGRES_DB=" + "postgres"})
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	if err = TestPool.Retry(func() error {
		var err error
		pgdb, err = sql.Open("postgres", fmt.Sprintf("postgres://postgres:postgres@%s:%s/%s?sslmode=disable", PostgresHost(), TestPostgresRsource.GetPort("5432/tcp"), "postgres"))
		if err != nil {
			return err
		}
		return pgdb.Ping()
	}); err != nil {
		t.Fatalf("Could not connect to postgres docker: %s", err)
	}

	t.Logf("Docker postgres started")
	return pgdb
}

func CountRows(db *sql.DB, table string) (count int, err error) {
	row := db.QueryRow("SELECT COUNT(*) FROM " + table)
	err = row.Scan(&count)
	return
}

func GetTables(db *sql.DB) ([]string, error) {
	var tables []string

	query := "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema';"

	rows, err := db.Query(query)
	if err != nil {
		return tables, err
	}

	for rows.Next() {
		var table string
		err := rows.Scan(&table)
		if err != nil {
			return tables, err
		}
		tables = append(tables, table)
	}

	return tables, err
}