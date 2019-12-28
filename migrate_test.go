package migrate

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

func TestMigrate(t *testing.T) {
	ctx := context.Background()
	m := Funcs{}
	m[1577566714] = Log(nil).Commands(
		`CREATE TABLE users (
			id SERIAL NOT NULL,
			email TEXT NOT NULL,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,
		`CREATE UNIQUE INDEX users_email_key ON users (email)`,
	)
	m[1577566893] = func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `ALTER TABLE users ADD COLUMN password_digest BYTEA`)
		return err
	}
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Exec(ctx, `SET search_path TO 'pg_temp'`)
	if err != nil {
		t.Fatal(err)
	}
	err = Migrate(ctx, conn, m)
	if err != nil {
		t.Fatal(err)
	}
	store := &store{conn: conn}
	err = store.insertUser(ctx, "aj@testing")
	if err != nil {
		t.Fatal(err)
	}
	err = store.insertUser(ctx, "aj@testing")
	if err == nil {
		t.Fatalf("expected pg error")
	}
	pgErr, ok := err.(*pgconn.PgError)
	if !ok {
		t.Fatalf("expected pg error: %t", err)
	}
	if pgErr.ConstraintName != "users_email_key" {
		t.Fatalf("expected pg error on users email key: %v", pgErr)
	}
}

type store struct {
	conn *pgx.Conn
}

func (s *store) insertUser(ctx context.Context, email string) error {
	_, err := s.conn.Exec(ctx, `INSERT INTO users (email, created_at, updated_at) VALUES ($1, now(), now())`, email)
	return err
}
