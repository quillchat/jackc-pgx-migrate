package migrate

import (
	"context"
	"fmt"
	"github.com/jackc/pgx"
	"sort"
)

// Helper for running multiple commands in sequence.
func Run(tx *pgx.Tx, qs []string) error {
	for _, q := range qs {
		_, err := tx.Exec(q)
		if err != nil {
			return err
		}
	}
	return nil
}

func New() *Migrate {
	return &Migrate{
		Table:  "migrations",
		Column: "mts",
		funcs:  make(map[int64]func(*pgx.Tx) error),
	}
}

type Migrate struct {
	Table  string
	Column string
	funcs  map[int64]func(*pgx.Tx) error
}

func (m *Migrate) Set(k int64, txFunc func(*pgx.Tx) error) {
	m.funcs[k] = txFunc
}

func (m *Migrate) Exec(conn *pgx.Conn) error {
	var ks []int64
	for k := range m.funcs {
		ks = append(ks, k)
	}
	sort.Slice(ks, func(i, j int) bool {
		return ks[i] < ks[j]
	})
	ctx := context.Background()
	q1 := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (`%s` BIGINT PRIMARY KEY)", m.Table, m.Column)
	_, err := conn.Exec(ctx, q1)
	if err != nil {
		return err
	}
	for _, k := range ks {
		q2 := fmt.Sprintf("SELECT 1 FROM `%s` WHERE `%s` = $1", m.Table, m.Column)
		var exists int
		err := conn.QueryRow(ctx, q2, k).Scan(&exists)
		if err != nil && err != pgx.ErrNoRows {
			return err
		}
		if exists == 1 {
			continue
		}
		tx, err := conn.Begin(ctx)
		if err != nil {
			return err
		}
		err = m.funcs[k](tx)
		if err != nil {
			tx.Rollback()
			return err
		}
		q3 := fmt.Sprintf("INSERT INTO `%s` (`%s`) VALUES ($1)", m.Table, m.Columns)
		_, err = tx.Exec(ctx, q3, k)
		if err != nil {
			tx.Rollback()
			return err
		}
		err = tx.Commit()
		if err != nil {
			return err
		}
	}
	return nil
}
