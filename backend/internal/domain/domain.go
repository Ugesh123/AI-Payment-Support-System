// Package domain holds the core models, enums and status state machines.
// It has no database or HTTP code, so the business rules are easy to test.
package domain

import (
	"fmt"
	"time"

	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/money"
)

// ---------- Enums ----------

type TxnStatus string

const (
	TxnSuccess TxnStatus = "SUCCESS"
	TxnFailed  TxnStatus = "FAILED"
	TxnPending TxnStatus = "PENDING"
)

type RefundStatus string

const (
	RefundRequested  RefundStatus = "REQUESTED"
	RefundProcessing RefundStatus = "PROCESSING"
	RefundCompleted  RefundStatus = "COMPLETED"
)

type TicketStatus string

const (
	TicketOpen       TicketStatus = "OPEN"
	TicketInProgress TicketStatus = "IN_PROGRESS"
	TicketClosed     TicketStatus = "CLOSED"
)

// PaymentMethod holds the API value ("Card", "Net Banking", ...).
type PaymentMethod string

// RefundReason holds the API value ("Duplicate Payment", ...).
type RefundReason string

// MySQL stores the enum NAME (e.g. NET_BANKING) while the JSON API uses
// the VALUE (e.g. "Net Banking"). These tables map between the two.
var paymentMethods = []struct{ Name, Value string }{
	{"UPI", "UPI"}, {"CARD", "Card"}, {"NET_BANKING", "Net Banking"}, {"WALLET", "Wallet"},
}

var refundReasons = []struct{ Name, Value string }{
	{"DUPLICATE_PAYMENT", "Duplicate Payment"}, {"WRONG_AMOUNT", "Wrong Amount"},
	{"ORDER_CANCELLED", "Order Cancelled"}, {"OTHER", "Other"},
}

func lookup(table []struct{ Name, Value string }, s string) (name, value string, ok bool) {
	for _, e := range table {
		if s == e.Name || s == e.Value {
			return e.Name, e.Value, true
		}
	}
	return "", "", false
}

// ParsePaymentMethod accepts either the API value or the DB name.
func ParsePaymentMethod(s string) (PaymentMethod, bool) {
	_, v, ok := lookup(paymentMethods, s)
	return PaymentMethod(v), ok
}

// DBName is what gets written to MySQL.
func (p PaymentMethod) DBName() string { n, _, _ := lookup(paymentMethods, string(p)); return n }

func ParseRefundReason(s string) (RefundReason, bool) {
	_, v, ok := lookup(refundReasons, s)
	return RefundReason(v), ok
}

func (r RefundReason) DBName() string { n, _, _ := lookup(refundReasons, string(r)); return n }

func ValidTxnStatus(s TxnStatus) bool {
	return s == TxnSuccess || s == TxnFailed || s == TxnPending
}
func ValidRefundStatus(s RefundStatus) bool {
	return s == RefundRequested || s == RefundProcessing || s == RefundCompleted
}
func ValidTicketStatus(s TicketStatus) bool {
	return s == TicketOpen || s == TicketInProgress || s == TicketClosed
}

// ---------- State machines ----------
// Each map lists the ONLY allowed forward moves; anything else is rejected.

var TxnTransitions = map[TxnStatus][]TxnStatus{
	TxnPending: {TxnSuccess, TxnFailed},
	TxnSuccess: {},
	TxnFailed:  {},
}

var RefundTransitions = map[RefundStatus][]RefundStatus{
	RefundRequested:  {RefundProcessing},
	RefundProcessing: {RefundCompleted},
	RefundCompleted:  {},
}

var TicketTransitions = map[TicketStatus][]TicketStatus{
	TicketOpen:       {TicketInProgress, TicketClosed},
	TicketInProgress: {TicketClosed},
	TicketClosed:     {},
}

// CheckTransition returns a user-facing error if from -> to is not allowed.
// kind is "Refund", "Ticket" or "Transaction" (used in the message).
func CheckTransition[S ~string](rules map[S][]S, kind string, from, to S) error {
	if from == to {
		return fmt.Errorf("%s is already in %s status.", kind, to)
	}
	for _, next := range rules[from] {
		if next == to {
			return nil
		}
	}
	return fmt.Errorf("Invalid %s status transition: %s -> %s", lower(kind), from, to)
}

func lower(s string) string {
	if s == "" {
		return s
	}
	b := []byte(s)
	if b[0] >= 'A' && b[0] <= 'Z' {
		b[0] += 'a' - 'A'
	}
	return string(b)
}

// ---------- Models (JSON shapes the React frontend expects) ----------

type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Transaction struct {
	ID            int64         `json:"id"`
	TransactionID string        `json:"transaction_id"`
	UserID        int64         `json:"-"`
	Amount        money.Amount  `json:"amount"`
	Status        TxnStatus     `json:"status"`
	FailureReason *string       `json:"failure_reason"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	ReceiverName  string        `json:"receiver_name"`
	CreatedAt     time.Time     `json:"created_at"`
}

type Refund struct {
	ID            int64        `json:"id"`
	TransactionID int64        `json:"transaction_id"` // internal transactions.id
	Amount        money.Amount `json:"amount"`
	Reason        RefundReason `json:"reason"`
	Status        RefundStatus `json:"status"`
	RequestedAt   time.Time    `json:"requested_at"`
	CompletedAt   *time.Time   `json:"completed_at"`
}

type Ticket struct {
	ID            int64        `json:"id"`
	TicketID      string       `json:"ticket_id"`
	UserID        int64        `json:"-"`
	TransactionID int64        `json:"transaction_id"`
	Subject       string       `json:"subject"`
	Description   string       `json:"description"`
	Status        TicketStatus `json:"status"`
	CreatedAt     time.Time    `json:"created_at"`
	ClosedAt      *time.Time   `json:"closed_at"`
}

type DashboardSummary struct {
	TotalTransactions  int          `json:"total_transactions"`
	SuccessfulPayments int          `json:"successful_payments"`
	FailedPayments     int          `json:"failed_payments"`
	PendingPayments    int          `json:"pending_payments"`
	TotalAmountPaid    money.Amount `json:"total_amount_paid"`
	ActiveRefunds      int          `json:"active_refunds"`
	CompletedRefunds   int          `json:"completed_refunds"`
}

type RecentTransaction struct {
	TransactionID string       `json:"transaction_id"`
	Receiver      string       `json:"receiver"`
	Amount        money.Amount `json:"amount"`
	Status        TxnStatus    `json:"status"`
}

type Dashboard struct {
	Summary            DashboardSummary    `json:"summary"`
	RecentTransactions []RecentTransaction `json:"recent_transactions"`
}

type AdminSummary struct {
	TotalUsers          int `json:"total_users"`
	TotalTransactions   int `json:"total_transactions"`
	PendingTransactions int `json:"pending_transactions"`
	ActiveRefunds       int `json:"active_refunds"`
	OpenTickets         int `json:"open_tickets"`
}
