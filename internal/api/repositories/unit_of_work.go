package repositories

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

type UoW struct {
	pool *pgxpool.Pool
}

func NewUoW(pool *pgxpool.Pool) *UoW {
	return &UoW{pool: pool}
}

// WithinTransaction выполняет функцию в транзакции.
// Гарантирует корректный rollback и не теряет ошибки.
func (u *UoW) WithinTransaction(ctx context.Context, fn func(context.Context, DBTX) error) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}

	// rollback должен выполниться, если транзакция не была успешно закоммичена
	rollbackErr := func() error {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			return err
		}
		return nil
	}

	// вызываем пользовательскую функцию
	if err := fn(ctx, tx); err != nil {
		if rbErr := rollbackErr(); rbErr != nil {
			// возвращаем ошибку в формате "основная + rollback"
			return fmt.Errorf("%w; rollback error: %v", err, rbErr)
		}
		return err
	}

	// commit
	if err := tx.Commit(ctx); err != nil {
		// если commit не удался, нужно сделать rollback
		if rbErr := rollbackErr(); rbErr != nil {
			return fmt.Errorf("commit error: %w; rollback error: %v", err, rbErr)
		}
		return err
	}

	return nil
}
