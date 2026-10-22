package store

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/domain"
)

func TestDSNFromURL(t *testing.T) {
	dsn, err := DSNFromURL("mysql://root:p%40ss@db.railway.internal:3306/railway")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"root:p@ss@tcp(db.railway.internal:3306)/railway", "parseTime=true"} {
		if !strings.Contains(dsn, want) {
			t.Errorf("dsn %q missing %q", dsn, want)
		}
	}
	if dsn, _ := DSNFromURL("mysql://u:p@localhost/app"); !strings.Contains(dsn, "tcp(localhost:3306)/app") {
		t.Errorf("default port not added: %s", dsn)
	}
	if _, err := DSNFromURL("postgres://u:p@h/db"); err == nil {
		t.Error("non-mysql URL should fail")
	}
	if dsn, _ := DSNFromURL("u:p@tcp(h:3306)/db"); dsn != "u:p@tcp(h:3306)/db" {
		t.Error("plain DSN should pass through unchanged")
	}
}

// Many admins clicking "Approve" at once: compare-and-set lets exactly one win.
func TestCompareAndSetUnderConcurrency(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()
	r := &domain.Refund{TransactionID: 1, Status: domain.RefundRequested}
	if err := m.CreateRefund(ctx, r); err != nil {
		t.Fatal(err)
	}

	var wins, stale atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := m.UpdateRefundStatus(ctx, r.ID, domain.RefundRequested, domain.RefundProcessing, nil)
			switch {
			case err == nil:
				wins.Add(1)
			case errors.Is(err, ErrStale):
				stale.Add(1)
			}
		}()
	}
	wg.Wait()
	if wins.Load() != 1 || stale.Load() != 49 {
		t.Fatalf("wins=%d stale=%d, want 1 and 49", wins.Load(), stale.Load())
	}
}

func TestUniqueRefundPerTransaction(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()
	if err := m.CreateRefund(ctx, &domain.Refund{TransactionID: 7}); err != nil {
		t.Fatal(err)
	}
	if err := m.CreateRefund(ctx, &domain.Refund{TransactionID: 7}); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("want ErrDuplicate, got %v", err)
	}
}
