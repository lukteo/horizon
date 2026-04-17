package testhelper_test

import (
	"context"
	"testing"

	"github.com/luketeo/horizon/internal/platform/testhelper"
)

func TestDB_SchemaIsMigrated(t *testing.T) {
	db := testhelper.DB(t)
	testhelper.Reset(t, db)

	var n int
	if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		t.Fatalf("expected users table to exist: %v", err)
	}
	if n != 0 {
		t.Errorf("expected empty users table after reset, got %d rows", n)
	}
}

func TestReset_ClearsInsertedRows(t *testing.T) {
	db := testhelper.DB(t)
	ctx := context.Background()
	testhelper.Reset(t, db)

	_, err := db.ExecContext(ctx,
		`INSERT INTO users (clerk_id, email) VALUES ('test_clerk_id', 'test@example.com')`)
	if err != nil {
		t.Fatalf("inserting: %v", err)
	}

	testhelper.Reset(t, db)

	var n int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		t.Fatalf("querying: %v", err)
	}
	if n != 0 {
		t.Errorf("expected reset to clear users, got %d rows", n)
	}
}
