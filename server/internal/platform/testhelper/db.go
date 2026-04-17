// Package testhelper provides shared utilities for integration tests that hit
// the real test Postgres instance. Each test is expected to start by calling
// Reset to truncate all domain tables.
package testhelper

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"

	_ "github.com/lib/pq" // pg driver
)

var (
	testDBOnce sync.Once
	testDB     *sql.DB
	testDBErr  error
)

// DB returns a shared *sql.DB connected to the test database specified by
// TEST_DATABASE_URL. Tests are skipped when the env var is unset so the suite
// remains runnable in environments without a local database.
func DB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping DB-backed test")
	}
	testDBOnce.Do(func() {
		testDB, testDBErr = sql.Open("postgres", dsn)
		if testDBErr == nil {
			testDBErr = testDB.Ping()
		}
	})
	if testDBErr != nil {
		t.Fatalf("opening test database: %v", testDBErr)
	}
	return testDB
}

// Reset truncates every domain table, restoring a blank slate. Tests are not
// parallel-safe against this harness: rely on sequential execution.
func Reset(t *testing.T, db *sql.DB) {
	t.Helper()
	const q = `TRUNCATE api_keys, organization_members, organizations, users RESTART IDENTITY CASCADE`
	if _, err := db.ExecContext(context.Background(), q); err != nil {
		t.Fatalf("truncating tables: %v", err)
	}
}
