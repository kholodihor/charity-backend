package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store struct {
	*Queries
	db *pgx.Conn
}

func NewStore(db *pgx.Conn) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

const (
	maxTxRetries     = 3
	txRetryBaseDelay = 10 * time.Millisecond
)

func isRetryableTxError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	switch pgErr.Code {
	case "40P01", // deadlock_detected
		"40001": // serialization_failure
		return true
	default:
		return false
	}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	var lastErr error

	for attempt := 0; attempt <= maxTxRetries; attempt++ {
		if attempt > 0 {
			// simple exponential backoff between retries
			delay := txRetryBaseDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		tx, err := store.db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			if isRetryableTxError(err) && attempt < maxTxRetries {
				lastErr = err
				continue
			}
			return err
		}

		q := New(tx)
		err = fn(q)
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
			}

			if isRetryableTxError(err) && attempt < maxTxRetries {
				lastErr = err
				continue
			}
			return err
		}

		if err = tx.Commit(ctx); err != nil {
			if isRetryableTxError(err) && attempt < maxTxRetries {
				lastErr = err
				continue
			}
			return err
		}

		// success
		return nil
	}

	if lastErr != nil {
		return lastErr
	}

	return fmt.Errorf("transaction failed after %d retries", maxTxRetries)
}

type DonationTxParams struct {
	UserID      pgtype.Int8 `json:"user_id"`
	GoalID      int64       `json:"goal_id"`
	Amount      int64       `json:"amount"`
	Currency    string      `json:"currency"`
	IsAnonymous bool        `json:"is_anonymous"`
}

type DonationTxResult struct {
	Donation Donation `json:"donation"`
}

func (store *Store) DonationTx(ctx context.Context, arg DonationTxParams) (DonationTxResult, error) {
	var result DonationTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		// lock the goal row for this donation
		if _, err := q.GetGoalForUpdate(ctx, arg.GoalID); err != nil {
			return err
		}

		// increment collected_amount atomically for the locked goal
		if _, err := q.AddToGoalCollectedAmount(ctx, AddToGoalCollectedAmountParams{
			ID:     arg.GoalID,
			Amount: arg.Amount,
		}); err != nil {
			return err
		}

		donation, err := q.CreateDonation(ctx, CreateDonationParams(arg))
		if err != nil {
			return err
		}

		result = DonationTxResult{Donation: donation}
		return nil
	})

	return result, err
}
