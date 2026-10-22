package store

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/domain"
)

// Memory is a thread-safe in-memory Store. It enforces the same unique
// constraints as the MySQL schema, so tests exercise the real rules.
type Memory struct {
	mu      sync.RWMutex
	seq     map[string]int64 // one auto-increment counter per table
	users   []domain.User
	txns    []domain.Transaction
	refunds []domain.Refund
	tickets []domain.Ticket
	now     func() time.Time
}

func NewMemory() *Memory {
	return &Memory{seq: map[string]int64{}, now: func() time.Time { return time.Now().UTC() }}
}

func (m *Memory) next(table string) int64 { m.seq[table]++; return m.seq[table] }

func (m *Memory) Close() error { return nil }

// ---------- Users ----------

func (m *Memory) CreateUser(_ context.Context, u *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, x := range m.users {
		if strings.EqualFold(x.Email, u.Email) {
			return ErrDuplicate
		}
	}
	u.ID, u.CreatedAt = m.next("users"), m.now()
	if u.Role == "" {
		u.Role = "user"
	}
	m.users = append(m.users, *u)
	return nil
}

func (m *Memory) UserByEmail(_ context.Context, email string) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, x := range m.users {
		if strings.EqualFold(x.Email, email) {
			u := x
			return &u, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) ListUsers(_ context.Context) ([]domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := append([]domain.User(nil), m.users...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out, nil
}

func (m *Memory) PromoteToAdmin(_ context.Context, email string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.users {
		if strings.EqualFold(m.users[i].Email, email) {
			m.users[i].Role = "admin"
			return nil
		}
	}
	return ErrNotFound
}

// ---------- Transactions ----------

func (m *Memory) CreateTransaction(_ context.Context, t *domain.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, x := range m.txns {
		if x.TransactionID == t.TransactionID {
			return ErrDuplicate
		}
	}
	t.ID, t.CreatedAt = m.next("transactions"), m.now()
	m.txns = append(m.txns, *t)
	return nil
}

func (m *Memory) TransactionByPublicID(_ context.Context, userID int64, publicID string) (*domain.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, x := range m.txns {
		if x.TransactionID == publicID && x.UserID == userID {
			t := x
			return &t, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) TransactionByID(_ context.Context, id int64) (*domain.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, x := range m.txns {
		if x.ID == id {
			t := x
			return &t, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) ListTransactions(_ context.Context, userID int64, limit int) ([]domain.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []domain.Transaction{}
	for i := len(m.txns) - 1; i >= 0; i-- { // newest first
		if userID == 0 || m.txns[i].UserID == userID {
			out = append(out, m.txns[i])
			if limit > 0 && len(out) == limit {
				break
			}
		}
	}
	return out, nil
}

func (m *Memory) UpdateTransactionStatus(_ context.Context, id int64, from, to domain.TxnStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.txns {
		if m.txns[i].ID == id {
			if m.txns[i].Status != from {
				return ErrStale
			}
			m.txns[i].Status = to
			return nil
		}
	}
	return ErrNotFound
}

// ---------- Refunds ----------

func (m *Memory) CreateRefund(_ context.Context, r *domain.Refund) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, x := range m.refunds { // UNIQUE(transaction_id)
		if x.TransactionID == r.TransactionID {
			return ErrDuplicate
		}
	}
	r.ID, r.RequestedAt = m.next("refunds"), m.now()
	m.refunds = append(m.refunds, *r)
	return nil
}

func (m *Memory) RefundByTransaction(_ context.Context, txnID int64) (*domain.Refund, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, x := range m.refunds {
		if x.TransactionID == txnID {
			r := x
			return &r, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) RefundByID(_ context.Context, id int64) (*domain.Refund, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, x := range m.refunds {
		if x.ID == id {
			r := x
			return &r, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) ListRefunds(_ context.Context) ([]domain.Refund, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []domain.Refund{}
	for i := len(m.refunds) - 1; i >= 0; i-- {
		out = append(out, m.refunds[i])
	}
	return out, nil
}

func (m *Memory) UpdateRefundStatus(_ context.Context, id int64, from, to domain.RefundStatus, completedAt *time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.refunds {
		if m.refunds[i].ID == id {
			if m.refunds[i].Status != from {
				return ErrStale
			}
			m.refunds[i].Status = to
			if completedAt != nil {
				m.refunds[i].CompletedAt = completedAt
			}
			return nil
		}
	}
	return ErrNotFound
}

// ---------- Tickets ----------

func (m *Memory) CreateTicket(_ context.Context, t *domain.Ticket) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, x := range m.tickets {
		if x.TicketID == t.TicketID {
			return ErrDuplicate
		}
	}
	t.ID, t.CreatedAt = m.next("tickets"), m.now()
	m.tickets = append(m.tickets, *t)
	return nil
}

func (m *Memory) TicketByPublicID(_ context.Context, userID int64, ticketID string) (*domain.Ticket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, x := range m.tickets {
		if x.TicketID == ticketID && x.UserID == userID {
			t := x
			return &t, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) TicketByID(_ context.Context, id int64) (*domain.Ticket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, x := range m.tickets {
		if x.ID == id {
			t := x
			return &t, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) ListTickets(_ context.Context, userID int64) ([]domain.Ticket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []domain.Ticket{}
	for i := len(m.tickets) - 1; i >= 0; i-- {
		if userID == 0 || m.tickets[i].UserID == userID {
			out = append(out, m.tickets[i])
		}
	}
	return out, nil
}

func (m *Memory) UpdateTicketStatus(_ context.Context, id int64, from, to domain.TicketStatus, closedAt *time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.tickets {
		if m.tickets[i].ID == id {
			if m.tickets[i].Status != from {
				return ErrStale
			}
			m.tickets[i].Status = to
			if closedAt != nil {
				m.tickets[i].ClosedAt = closedAt
			}
			return nil
		}
	}
	return ErrNotFound
}

// ---------- Aggregates ----------

func (m *Memory) UserSummary(_ context.Context, userID int64) (domain.DashboardSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var s domain.DashboardSummary
	mine := map[int64]bool{}
	for _, t := range m.txns {
		if t.UserID != userID {
			continue
		}
		mine[t.ID] = true
		s.TotalTransactions++
		switch t.Status {
		case domain.TxnSuccess:
			s.SuccessfulPayments++
			s.TotalAmountPaid += t.Amount
		case domain.TxnFailed:
			s.FailedPayments++
		case domain.TxnPending:
			s.PendingPayments++
		}
	}
	for _, r := range m.refunds {
		if !mine[r.TransactionID] {
			continue
		}
		if r.Status == domain.RefundCompleted {
			s.CompletedRefunds++
		} else {
			s.ActiveRefunds++
		}
	}
	return s, nil
}

func (m *Memory) AdminSummary(_ context.Context) (domain.AdminSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := domain.AdminSummary{TotalUsers: len(m.users), TotalTransactions: len(m.txns)}
	for _, t := range m.txns {
		if t.Status == domain.TxnPending {
			s.PendingTransactions++
		}
	}
	for _, r := range m.refunds {
		if r.Status != domain.RefundCompleted {
			s.ActiveRefunds++
		}
	}
	for _, t := range m.tickets {
		if t.Status != domain.TicketClosed {
			s.OpenTickets++
		}
	}
	return s, nil
}
