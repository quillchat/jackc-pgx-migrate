package migrate

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v4"
)

type Funcs map[int64]func(context.Context, pgx.Tx) error

func Commands(qs ...string) func(context.Context, pgx.Tx) error {
	return func(ctx context.Context, tx pgx.Tx) error {
		return Run(ctx, tx, qs)
	}
}

// Helper for running multiple commands in sequence.
func Run(ctx context.Context, tx pgx.Tx, qs []string) error {
	for _, q := range qs {
		_, err := tx.Exec(ctx, q)
		if err != nil {
			return err
		}
	}
	return nil
}

func Migrate(ctx context.Context, conn *pgx.Conn, funcs Funcs, logger *Logger) error {
	var ks []int64
	for k := range funcs {
		ks = append(ks, k)
	}
	sort.Slice(ks, func(i, j int) bool {
		return ks[i] < ks[j]
	})
	_, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS migrations (mts BIGINT PRIMARY KEY)`)
	if err != nil {
		return err
	}
	for _, k := range ks {
		var exists int
		err := conn.QueryRow(ctx, `SELECT 1 FROM migrations WHERE mts = $1`, k).Scan(&exists)
		if err != nil && err != pgx.ErrNoRows {
			return err
		}
		if exists == 1 {
			continue
		}

		logger.Printf("Executing migration: %d", k)
		start := time.Now()
		tx, err := conn.Begin(ctx)
		if err != nil {
			return err
		}
		err = funcs[k](ctx, tx)
		if err != nil {
			tx.Rollback(ctx)
			return &Error{Err: err, Key: k}
		}
		_, err = tx.Exec(ctx, `INSERT INTO migrations (mts) VALUES ($1)`, k)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
		err = tx.Commit(ctx)
		if err != nil {
			return err
		}
		logger.Printf("Executed migration: %d, took %0.5f seconds", k, time.Now().Sub(start).Seconds())
	}
	return nil
}

type Error struct {
	Err error
	Key int64
}

func (e *Error) Error() string {
	return fmt.Sprintf("migrate %d: %v", e.Key, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func Log(printer Printer) *Logger {
	return &Logger{Printer: printer}
}

type Logger struct {
	Printer Printer
}

func (l *Logger) Commands(qs ...string) func(context.Context, pgx.Tx) error {
	return func(ctx context.Context, tx pgx.Tx) error {
		l.Printf("BEGIN")
		err := l.Run(ctx, tx, qs)
		if err != nil {
			l.Printf("ROLLBACK")
			return err
		}
		l.Printf("COMMIT")
		return nil
	}
}

func (l *Logger) Run(ctx context.Context, tx pgx.Tx, qs []string) error {
	for _, q := range qs {
		l.Printf(q)
		_, err := tx.Exec(ctx, q)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *Logger) Printf(s string, v ...interface{}) {
	if l.Printer != nil {
		l.Printer.Printf(s, v...)
	}
}

type Printer interface {
	Printf(string, ...interface{})
}
