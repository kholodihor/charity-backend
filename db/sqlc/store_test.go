package db

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL or DATABASE_URL must be set for integration tests")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Clean tables that are relevant for these tests
	_, err = conn.Exec(ctx, "DELETE FROM donations")
	if err != nil {
		conn.Close(ctx)
		t.Fatalf("failed to clean donations table: %v", err)
	}
	_, err = conn.Exec(ctx, "DELETE FROM goals")
	if err != nil {
		conn.Close(ctx)
		t.Fatalf("failed to clean goals table: %v", err)
	}

	store := NewStore(conn)
	return store
}

func TestDonationTxUpdatesCollectedAmount(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	goal, err := store.CreateGoal(ctx, CreateGoalParams{})
	if err != nil {
		t.Fatalf("failed to create goal: %v", err)
	}

	amount := int64(100)
	_, err = store.DonationTx(ctx, DonationTxParams{
		GoalID:      goal.ID,
		Amount:      amount,
		Currency:    "USD",
		IsAnonymous: true,
	})
	if err != nil {
		t.Fatalf("DonationTx failed: %v", err)
	}

	updated, err := store.GetGoal(ctx, goal.ID)
	if err != nil {
		t.Fatalf("failed to fetch updated goal: %v", err)
	}

	if updated.CollectedAmount != amount {
		t.Fatalf("unexpected collected_amount: got %d, want %d", updated.CollectedAmount, amount)
	}
}

func TestDonationTxConcurrent(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	goal, err := store.CreateGoal(ctx, CreateGoalParams{})
	if err != nil {
		t.Fatalf("failed to create goal: %v", err)
	}

	const (
		workers = 5
		amount  = int64(50)
	)

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			if _, err := store.DonationTx(ctx, DonationTxParams{
				GoalID:      goal.ID,
				Amount:      amount,
				Currency:    "USD",
				IsAnonymous: true,
			}); err != nil {
				// Best-effort logging inside tests; we fail at the end if needed
				t.Errorf("DonationTx failed in goroutine: %v", err)
			}
		}()
	}

	wg.Wait()

	updated, err := store.GetGoal(ctx, goal.ID)
	if err != nil {
		t.Fatalf("failed to fetch updated goal: %v", err)
	}

	want := int64(workers) * amount
	if updated.CollectedAmount != want {
		t.Fatalf("unexpected collected_amount after concurrent donations: got %d, want %d", updated.CollectedAmount, want)
	}
}
