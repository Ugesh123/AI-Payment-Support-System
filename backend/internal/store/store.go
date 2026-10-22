// Package store defines the data-access interface and two implementations:
// MySQL (production) and in-memory (tests and quick local demos).
package store

import (
	"context"
	"errors"
	"time"

	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/domain"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrDuplicate = errors.New("duplicate")
	// ErrStale means a compare-and-set status update lost a race:
	// the row's status changed between the read and the write.
	ErrStale = errors.New("status changed by another request")
)

type Store interface {
	// Users
	CreateUser(ctx context.Context, u *domain.User) error
	UserByEmail(ctx context.Context, email string) (*domain.User, error)
	ListUsers(ctx context.Context) ([]domain.User, error)
	PromoteToAdmin(ctx context.Context, email string) error

	// Transactions
	CreateTransaction(ctx context.Context, t *domain.Transaction) error
	TransactionByPublicID(ctx context.Context, userID int64, publicID string) (*domain.Transaction, error)
	// ListTransactions returns newest first. userID 0 = all users; limit 0 = no limit.
	ListTransactions(ctx context.Context, userID int64, limit int) ([]domain.Transaction, error)
	TransactionByID(ctx context.Context, id int64) (*domain.Transaction, error)
	UpdateTransactionStatus(ctx context.Context, id int64, from, to domain.TxnStatus) error

	// Refunds
	CreateRefund(ctx context.Context, r *domain.Refund) error
	RefundByTransaction(ctx context.Context, txnID int64) (*domain.Refund, error)
	RefundByID(ctx context.Context, id int64) (*domain.Refund, error)
	ListRefunds(ctx context.Context) ([]domain.Refund, error)
	UpdateRefundStatus(ctx context.Context, id int64, from, to domain.RefundStatus, completedAt *time.Time) error

	// Tickets
	CreateTicket(ctx context.Context, t *domain.Ticket) error
	TicketByPublicID(ctx context.Context, userID int64, ticketID string) (*domain.Ticket, error)
	TicketByID(ctx context.Context, id int64) (*domain.Ticket, error)
	// ListTickets returns newest first. userID 0 = all users.
	ListTickets(ctx context.Context, userID int64) ([]domain.Ticket, error)
	UpdateTicketStatus(ctx context.Context, id int64, from, to domain.TicketStatus, closedAt *time.Time) error

	// Aggregates
	UserSummary(ctx context.Context, userID int64) (domain.DashboardSummary, error)
	AdminSummary(ctx context.Context) (domain.AdminSummary, error)

	Close() error
}
