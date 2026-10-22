// Package service holds the business logic. Handlers call these methods;
// these methods call the Store.
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	mrand "math/rand/v2"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/auth"
	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/domain"
	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/money"
	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/receipt"
	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/store"
)

// Error is a business error with the HTTP status it should map to.
type Error struct {
	Status int
	Detail string
}

func (e *Error) Error() string { return e.Detail }

func fail(status int, format string, args ...any) error {
	return &Error{Status: status, Detail: fmt.Sprintf(format, args...)}
}

var (
	MinAmount = money.FromRupees(1)
	MaxAmount = money.FromRupees(100000)
)

// PaymentSimulator stands in for a real payment gateway.
type PaymentSimulator func() (domain.TxnStatus, *string)

var failureReasons = []string{"Insufficient Balance", "Bank Server Timeout", "Card Declined", "Incorrect OTP"}

// DefaultPaymentSimulator: 70% SUCCESS, 20% FAILED, 10% PENDING.
func DefaultPaymentSimulator() (domain.TxnStatus, *string) {
	n := mrand.IntN(100)
	switch {
	case n < 70:
		return domain.TxnSuccess, nil
	case n < 90:
		r := failureReasons[mrand.IntN(len(failureReasons))]
		return domain.TxnFailed, &r
	default:
		return domain.TxnPending, nil
	}
}

type Service struct {
	Store  store.Store
	Tokens *auth.TokenManager
	Pay    PaymentSimulator
	Now    func() time.Time
}

func New(st store.Store, tokens *auth.TokenManager) *Service {
	return &Service{Store: st, Tokens: tokens, Pay: DefaultPaymentSimulator,
		Now: func() time.Time { return time.Now().UTC() }}
}

// publicID returns e.g. "TXN-9F2A4C8B".
func publicID(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return prefix + "-" + strings.ToUpper(hex.EncodeToString(b))
}

// ---------- Auth ----------

type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func validEmail(s string) bool {
	a, err := mail.ParseAddress(s)
	return err == nil && a.Address == s && strings.Contains(s, ".")
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*domain.User, error) {
	in.Name, in.Email = strings.TrimSpace(in.Name), strings.TrimSpace(in.Email)
	switch {
	case in.Name == "":
		return nil, fail(http.StatusUnprocessableEntity, "Name is required")
	case !validEmail(in.Email):
		return nil, fail(http.StatusUnprocessableEntity, "A valid email address is required")
	case in.Password == "":
		return nil, fail(http.StatusUnprocessableEntity, "Password is required")
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}
	u := &domain.User{Name: in.Name, Email: in.Email, PasswordHash: hash, Role: "user"}
	if err := s.Store.CreateUser(ctx, u); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			return nil, fail(http.StatusBadRequest, "Email already registered")
		}
		return nil, err
	}
	return u, nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (string, error) {
	u, err := s.Store.UserByEmail(ctx, strings.TrimSpace(in.Email))
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return "", err
	}
	// Same message for "no such user" and "wrong password", so the API
	// doesn't reveal which emails are registered.
	if u == nil || !auth.VerifyPassword(in.Password, u.PasswordHash) {
		return "", fail(http.StatusUnauthorized, "Invalid email or password")
	}
	return s.Tokens.Issue(u.Email)
}

// UserFromToken resolves a bearer token to a user.
func (s *Service) UserFromToken(ctx context.Context, token string) (*domain.User, error) {
	email, err := s.Tokens.Parse(token)
	if err != nil {
		return nil, fail(http.StatusUnauthorized, "Invalid or expired token")
	}
	u, err := s.Store.UserByEmail(ctx, email)
	if errors.Is(err, store.ErrNotFound) {
		return nil, fail(http.StatusUnauthorized, "User not found")
	}
	return u, err
}

// EnsureAdmin creates (if needed) and promotes an admin account at startup.
func (s *Service) EnsureAdmin(ctx context.Context, email, password string) error {
	if _, err := s.Store.UserByEmail(ctx, email); errors.Is(err, store.ErrNotFound) {
		if _, err := s.Register(ctx, RegisterInput{Name: "Admin", Email: email, Password: password}); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return s.Store.PromoteToAdmin(ctx, email)
}

// ---------- Transactions ----------

type TransactionInput struct {
	Amount        *money.Amount `json:"amount"`
	ReceiverName  string        `json:"receiver_name"`
	PaymentMethod string        `json:"payment_method"`
}

func (s *Service) CreateTransaction(ctx context.Context, u *domain.User, in TransactionInput) (*domain.Transaction, error) {
	in.ReceiverName = strings.TrimSpace(in.ReceiverName)
	if in.Amount == nil {
		return nil, fail(http.StatusUnprocessableEntity, "Amount is required")
	}
	if *in.Amount <= 0 {
		return nil, fail(http.StatusUnprocessableEntity, "Amount must be greater than zero")
	}
	if *in.Amount < MinAmount || *in.Amount > MaxAmount {
		return nil, fail(http.StatusUnprocessableEntity, "Amount must be between %s and %s", MinAmount, MaxAmount)
	}
	if in.ReceiverName == "" || utf8.RuneCountInString(in.ReceiverName) > 100 {
		return nil, fail(http.StatusUnprocessableEntity, "Receiver name is required (max 100 characters)")
	}
	pm, ok := domain.ParsePaymentMethod(in.PaymentMethod)
	if !ok {
		return nil, fail(http.StatusUnprocessableEntity, "Payment method must be one of: UPI, Card, Net Banking, Wallet")
	}

	status, reason := s.Pay()
	t := &domain.Transaction{
		UserID: u.ID, Amount: *in.Amount, Status: status, FailureReason: reason,
		PaymentMethod: pm, ReceiverName: in.ReceiverName,
	}
	// Retry on the (very unlikely) chance of an ID collision.
	for attempt := 0; attempt < 3; attempt++ {
		t.TransactionID = publicID("TXN")
		err := s.Store.CreateTransaction(ctx, t)
		if !errors.Is(err, store.ErrDuplicate) {
			return t, err
		}
	}
	return nil, errors.New("could not generate a unique transaction id")
}

func (s *Service) ListTransactions(ctx context.Context, u *domain.User) ([]domain.Transaction, error) {
	return s.Store.ListTransactions(ctx, u.ID, 0)
}

func (s *Service) GetTransaction(ctx context.Context, u *domain.User, publicID string) (*domain.Transaction, error) {
	t, err := s.Store.TransactionByPublicID(ctx, u.ID, publicID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, fail(http.StatusNotFound, "Transaction not found")
	}
	return t, err
}

// ---------- Refunds ----------

type RefundInput struct {
	TransactionID string `json:"transaction_id"`
	RefundReason  string `json:"refund_reason"`
}

func (s *Service) RequestRefund(ctx context.Context, u *domain.User, in RefundInput) (*domain.Refund, error) {
	reason, ok := domain.ParseRefundReason(in.RefundReason)
	if !ok {
		return nil, fail(http.StatusUnprocessableEntity,
			"Refund reason must be one of: Duplicate Payment, Wrong Amount, Order Cancelled, Other")
	}
	t, err := s.GetTransaction(ctx, u, in.TransactionID)
	if err != nil {
		return nil, err
	}
	if t.Status != domain.TxnSuccess {
		return nil, fail(http.StatusBadRequest, "Refund can only be requested for successful transactions")
	}
	r := &domain.Refund{TransactionID: t.ID, Amount: t.Amount, Reason: reason, Status: domain.RefundRequested}
	if err := s.Store.CreateRefund(ctx, r); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			return nil, fail(http.StatusBadRequest, "A refund has already been requested for this transaction")
		}
		return nil, err
	}
	return r, nil
}

func (s *Service) GetRefund(ctx context.Context, u *domain.User, txnPublicID string) (*domain.Refund, error) {
	t, err := s.Store.TransactionByPublicID(ctx, u.ID, txnPublicID)
	if err == nil {
		var r *domain.Refund
		if r, err = s.Store.RefundByTransaction(ctx, t.ID); err == nil {
			return r, nil
		}
	}
	if errors.Is(err, store.ErrNotFound) {
		return nil, fail(http.StatusNotFound, "Refund not found")
	}
	return nil, err
}

func (s *Service) ListAllRefunds(ctx context.Context) ([]domain.Refund, error) {
	return s.Store.ListRefunds(ctx)
}

type StatusInput struct {
	Status string `json:"status"`
}

const staleMsg = "%s was updated by another request. Please refresh and try again."

func (s *Service) UpdateRefundStatus(ctx context.Context, id int64, in StatusInput) (*domain.Refund, error) {
	to := domain.RefundStatus(in.Status)
	if !domain.ValidRefundStatus(to) {
		return nil, fail(http.StatusUnprocessableEntity, "Status must be one of: REQUESTED, PROCESSING, COMPLETED")
	}
	r, err := s.Store.RefundByID(ctx, id)
	if errors.Is(err, store.ErrNotFound) {
		return nil, fail(http.StatusNotFound, "Refund not found")
	}
	if err != nil {
		return nil, err
	}
	if err := domain.CheckTransition(domain.RefundTransitions, "Refund", r.Status, to); err != nil {
		return nil, fail(http.StatusBadRequest, "%s", err.Error())
	}
	var completedAt *time.Time
	if to == domain.RefundCompleted {
		now := s.Now().Truncate(time.Second)
		completedAt = &now
	}
	if err := s.Store.UpdateRefundStatus(ctx, id, r.Status, to, completedAt); err != nil {
		if errors.Is(err, store.ErrStale) {
			return nil, fail(http.StatusConflict, staleMsg, "Refund")
		}
		return nil, err
	}
	return s.Store.RefundByID(ctx, id)
}

// ---------- Tickets ----------

type TicketInput struct {
	TransactionID string `json:"transaction_id"`
	Subject       string `json:"subject"`
	Description   string `json:"description"`
}

func (s *Service) CreateTicket(ctx context.Context, u *domain.User, in TicketInput) (*domain.Ticket, error) {
	in.Subject, in.Description = strings.TrimSpace(in.Subject), strings.TrimSpace(in.Description)
	if in.Subject == "" || utf8.RuneCountInString(in.Subject) > 150 {
		return nil, fail(http.StatusUnprocessableEntity, "Subject is required (max 150 characters)")
	}
	if in.Description == "" {
		return nil, fail(http.StatusUnprocessableEntity, "Description is required")
	}
	t, err := s.GetTransaction(ctx, u, in.TransactionID)
	if err != nil {
		return nil, err
	}
	tk := &domain.Ticket{UserID: u.ID, TransactionID: t.ID, Subject: in.Subject,
		Description: in.Description, Status: domain.TicketOpen}
	for attempt := 0; attempt < 3; attempt++ {
		tk.TicketID = publicID("TKT")
		err := s.Store.CreateTicket(ctx, tk)
		if !errors.Is(err, store.ErrDuplicate) {
			return tk, err
		}
	}
	return nil, errors.New("could not generate a unique ticket id")
}

func (s *Service) ListMyTickets(ctx context.Context, u *domain.User) ([]domain.Ticket, error) {
	return s.Store.ListTickets(ctx, u.ID)
}

func (s *Service) GetTicket(ctx context.Context, u *domain.User, ticketID string) (*domain.Ticket, error) {
	t, err := s.Store.TicketByPublicID(ctx, u.ID, ticketID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, fail(http.StatusNotFound, "Ticket not found")
	}
	return t, err
}

func (s *Service) ListAllTickets(ctx context.Context) ([]domain.Ticket, error) {
	return s.Store.ListTickets(ctx, 0)
}

func (s *Service) UpdateTicketStatus(ctx context.Context, id int64, in StatusInput) (*domain.Ticket, error) {
	to := domain.TicketStatus(in.Status)
	if !domain.ValidTicketStatus(to) {
		return nil, fail(http.StatusUnprocessableEntity, "Status must be one of: OPEN, IN_PROGRESS, CLOSED")
	}
	t, err := s.Store.TicketByID(ctx, id)
	if errors.Is(err, store.ErrNotFound) {
		return nil, fail(http.StatusNotFound, "Ticket not found")
	}
	if err != nil {
		return nil, err
	}
	if err := domain.CheckTransition(domain.TicketTransitions, "Ticket", t.Status, to); err != nil {
		return nil, fail(http.StatusBadRequest, "%s", err.Error())
	}
	var closedAt *time.Time
	if to == domain.TicketClosed {
		now := s.Now().Truncate(time.Second)
		closedAt = &now
	}
	if err := s.Store.UpdateTicketStatus(ctx, id, t.Status, to, closedAt); err != nil {
		if errors.Is(err, store.ErrStale) {
			return nil, fail(http.StatusConflict, staleMsg, "Ticket")
		}
		return nil, err
	}
	return s.Store.TicketByID(ctx, id)
}

// ---------- Admin ----------

func (s *Service) ListUsers(ctx context.Context) ([]domain.User, error) {
	return s.Store.ListUsers(ctx)
}

func (s *Service) ListAllTransactions(ctx context.Context) ([]domain.Transaction, error) {
	return s.Store.ListTransactions(ctx, 0, 0)
}

func (s *Service) UpdateTransactionStatus(ctx context.Context, id int64, in StatusInput) (*domain.Transaction, error) {
	to := domain.TxnStatus(in.Status)
	if !domain.ValidTxnStatus(to) {
		return nil, fail(http.StatusUnprocessableEntity, "Status must be one of: SUCCESS, FAILED, PENDING")
	}
	t, err := s.Store.TransactionByID(ctx, id)
	if errors.Is(err, store.ErrNotFound) {
		return nil, fail(http.StatusNotFound, "Transaction not found")
	}
	if err != nil {
		return nil, err
	}
	if err := domain.CheckTransition(domain.TxnTransitions, "Transaction", t.Status, to); err != nil {
		return nil, fail(http.StatusBadRequest, "%s", err.Error())
	}
	if err := s.Store.UpdateTransactionStatus(ctx, id, t.Status, to); err != nil {
		if errors.Is(err, store.ErrStale) {
			return nil, fail(http.StatusConflict, staleMsg, "Transaction")
		}
		return nil, err
	}
	return s.Store.TransactionByID(ctx, id)
}

func (s *Service) AdminDashboard(ctx context.Context) (domain.AdminSummary, error) {
	return s.Store.AdminSummary(ctx)
}

// ---------- Dashboard, receipt, assistant ----------

func (s *Service) Dashboard(ctx context.Context, u *domain.User) (*domain.Dashboard, error) {
	sum, err := s.Store.UserSummary(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	recent, err := s.Store.ListTransactions(ctx, u.ID, 5)
	if err != nil {
		return nil, err
	}
	d := &domain.Dashboard{Summary: sum, RecentTransactions: make([]domain.RecentTransaction, 0, len(recent))}
	for _, t := range recent {
		d.RecentTransactions = append(d.RecentTransactions, domain.RecentTransaction{
			TransactionID: t.TransactionID, Receiver: t.ReceiverName, Amount: t.Amount, Status: t.Status,
		})
	}
	return d, nil
}

func (s *Service) Receipt(ctx context.Context, u *domain.User, txnPublicID string) ([]byte, error) {
	t, err := s.GetTransaction(ctx, u, txnPublicID)
	if err != nil {
		return nil, err
	}
	if t.Status != domain.TxnSuccess {
		return nil, fail(http.StatusBadRequest, "Receipt can only be generated for successful transactions.")
	}
	return receipt.Build(t), nil
}

type AssistantInput struct {
	TransactionID string `json:"transaction_id"`
	Question      string `json:"question"`
}

// Assistant is currently switched off and returns a maintenance message.
func (s *Service) Assistant(ctx context.Context, u *domain.User, in AssistantInput) (string, error) {
	t, err := s.GetTransaction(ctx, u, in.TransactionID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`
🚧 AI Assistant is Currently Under Maintenance

Thank you for using the AI Payment Support System.

Our AI Assistant is temporarily unavailable while we upgrade and improve its capabilities.

Transaction ID: %s
Payment Status: %s

For immediate assistance, you can:
• Check your transaction details.
• Request a refund (if eligible).
• Raise a support ticket.
• Contact customer support.

We appreciate your patience and will bring the AI Assistant back soon.
`, t.TransactionID, t.Status), nil
}
