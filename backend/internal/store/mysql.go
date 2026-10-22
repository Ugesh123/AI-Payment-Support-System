package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/domain"
)

// MySQL implements Store with database/sql and plain, parameterised SQL
// (no ORM), so every query is visible and safe from SQL injection.
type MySQL struct {
	db *sql.DB
	// enumValues is true when the enum columns store API values
	// ("Net Banking") as in docs/database_schema.sql, and false when they
	// store names ("NET_BANKING") as created by the migration below.
	enumValues bool
}

// DSNFromURL converts "mysql://user:pass@host:3306/db" (Railway style)
// into a go-sql-driver DSN. A ready-made DSN is returned unchanged.
func DSNFromURL(raw string) (string, error) {
	if !strings.Contains(raw, "://") {
		return raw, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid DATABASE_URL: %w", err)
	}
	if !strings.HasPrefix(u.Scheme, "mysql") {
		return "", fmt.Errorf("DATABASE_URL must be a mysql:// URL, got %q", u.Scheme)
	}
	cfg := mysql.NewConfig()
	cfg.User = u.User.Username()
	cfg.Passwd, _ = u.User.Password()
	cfg.Net = "tcp"
	cfg.Addr = u.Host
	if u.Port() == "" {
		cfg.Addr = u.Host + ":3306"
	}
	cfg.DBName = strings.TrimPrefix(u.Path, "/")
	cfg.ParseTime = true
	cfg.Loc = time.UTC
	return cfg.FormatDSN(), nil
}

// OpenMySQL connects, configures the pool, creates missing tables and
// detects how enum columns are stored.
func OpenMySQL(ctx context.Context, databaseURL string) (*MySQL, error) {
	dsn, err := DSNFromURL(databaseURL)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot reach MySQL: %w", err)
	}
	s := &MySQL{db: db}
	if err := s.migrate(ctx); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.detectEnumStyle(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *MySQL) Close() error { return s.db.Close() }

// Tables are created on startup only if they don't exist yet.
var schema = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(150) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		role VARCHAR(20) DEFAULT 'user',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS transactions (
		id INT AUTO_INCREMENT PRIMARY KEY,
		transaction_id VARCHAR(50) NOT NULL UNIQUE,
		user_id INT NOT NULL,
		amount NUMERIC(10,2) NOT NULL,
		status ENUM('SUCCESS','FAILED','PENDING') NOT NULL,
		failure_reason VARCHAR(255) NULL,
		payment_method ENUM('UPI','CARD','NET_BANKING','WALLET') NOT NULL,
		receiver_name VARCHAR(100) NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_txn_user_created (user_id, created_at),
		FOREIGN KEY (user_id) REFERENCES users(id)
	)`,
	`CREATE TABLE IF NOT EXISTS refunds (
		id INT AUTO_INCREMENT PRIMARY KEY,
		transaction_id INT NOT NULL UNIQUE,
		amount NUMERIC(10,2) NOT NULL,
		reason ENUM('DUPLICATE_PAYMENT','WRONG_AMOUNT','ORDER_CANCELLED','OTHER') NOT NULL,
		status ENUM('REQUESTED','PROCESSING','COMPLETED') NOT NULL DEFAULT 'REQUESTED',
		requested_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME NULL,
		FOREIGN KEY (transaction_id) REFERENCES transactions(id)
	)`,
	`CREATE TABLE IF NOT EXISTS support_tickets (
		id INT AUTO_INCREMENT PRIMARY KEY,
		ticket_id VARCHAR(50) NOT NULL UNIQUE,
		user_id INT NOT NULL,
		transaction_id INT NOT NULL,
		subject VARCHAR(150) NOT NULL,
		description TEXT NOT NULL,
		status ENUM('OPEN','IN_PROGRESS','CLOSED') NOT NULL DEFAULT 'OPEN',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		closed_at DATETIME NULL,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (transaction_id) REFERENCES transactions(id)
	)`,
}

func (s *MySQL) migrate(ctx context.Context) error {
	for _, stmt := range schema {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

func (s *MySQL) detectEnumStyle(ctx context.Context) error {
	var colType string
	err := s.db.QueryRowContext(ctx, `
		SELECT COLUMN_TYPE FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'transactions' AND COLUMN_NAME = 'payment_method'`,
	).Scan(&colType)
	if err != nil {
		return fmt.Errorf("detect enum style: %w", err)
	}
	s.enumValues = strings.Contains(colType, "'Net Banking'")
	return nil
}

func (s *MySQL) pm(p domain.PaymentMethod) string {
	if s.enumValues {
		return string(p)
	}
	return p.DBName()
}

func (s *MySQL) reason(r domain.RefundReason) string {
	if s.enumValues {
		return string(r)
	}
	return r.DBName()
}

func isDuplicate(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == 1062
}

// casResult turns "0 rows updated" into ErrNotFound or ErrStale.
func (s *MySQL) casResult(ctx context.Context, res sql.Result, table string, id int64) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 1 {
		return nil
	}
	var exists int
	err = s.db.QueryRowContext(ctx, "SELECT 1 FROM "+table+" WHERE id = ?", id).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return ErrStale
}

// ---------- Users ----------

const userCols = "id, name, email, password_hash, COALESCE(role,'user'), created_at"

func scanUser(row interface{ Scan(...any) error }) (*domain.User, error) {
	var u domain.User
	var created sql.NullTime
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	u.CreatedAt = created.Time
	return &u, nil
}

func (s *MySQL) CreateUser(ctx context.Context, u *domain.User) error {
	if u.Role == "" {
		u.Role = "user"
	}
	u.CreatedAt = time.Now().UTC().Truncate(time.Second)
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO users (name, email, password_hash, role, created_at) VALUES (?, ?, ?, ?, ?)",
		u.Name, u.Email, u.PasswordHash, u.Role, u.CreatedAt)
	if isDuplicate(err) {
		return ErrDuplicate
	}
	if err != nil {
		return err
	}
	u.ID, err = res.LastInsertId()
	return err
}

func (s *MySQL) UserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return scanUser(s.db.QueryRowContext(ctx, "SELECT "+userCols+" FROM users WHERE email = ?", email))
}

func (s *MySQL) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+userCols+" FROM users ORDER BY created_at DESC, id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *u)
	}
	return out, rows.Err()
}

func (s *MySQL) PromoteToAdmin(ctx context.Context, email string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE users SET role = 'admin' WHERE email = ?", email)
	return err
}

// ---------- Transactions ----------

const txnCols = "id, transaction_id, user_id, amount, status, failure_reason, payment_method, receiver_name, created_at"

func scanTxn(row interface{ Scan(...any) error }) (*domain.Transaction, error) {
	var t domain.Transaction
	var reason sql.NullString
	var pm string
	var created sql.NullTime
	if err := row.Scan(&t.ID, &t.TransactionID, &t.UserID, &t.Amount, &t.Status, &reason, &pm, &t.ReceiverName, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if reason.Valid {
		t.FailureReason = &reason.String
	}
	t.PaymentMethod, _ = domain.ParsePaymentMethod(pm)
	t.CreatedAt = created.Time
	return &t, nil
}

func (s *MySQL) CreateTransaction(ctx context.Context, t *domain.Transaction) error {
	t.CreatedAt = time.Now().UTC().Truncate(time.Second)
	res, err := s.db.ExecContext(ctx, `INSERT INTO transactions
		(transaction_id, user_id, amount, status, failure_reason, payment_method, receiver_name, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.TransactionID, t.UserID, t.Amount, t.Status, t.FailureReason, s.pm(t.PaymentMethod), t.ReceiverName, t.CreatedAt)
	if isDuplicate(err) {
		return ErrDuplicate
	}
	if err != nil {
		return err
	}
	t.ID, err = res.LastInsertId()
	return err
}

func (s *MySQL) TransactionByPublicID(ctx context.Context, userID int64, publicID string) (*domain.Transaction, error) {
	return scanTxn(s.db.QueryRowContext(ctx,
		"SELECT "+txnCols+" FROM transactions WHERE transaction_id = ? AND user_id = ?", publicID, userID))
}

func (s *MySQL) TransactionByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	return scanTxn(s.db.QueryRowContext(ctx, "SELECT "+txnCols+" FROM transactions WHERE id = ?", id))
}

func (s *MySQL) ListTransactions(ctx context.Context, userID int64, limit int) ([]domain.Transaction, error) {
	q := "SELECT " + txnCols + " FROM transactions"
	args := []any{}
	if userID != 0 {
		q += " WHERE user_id = ?"
		args = append(args, userID)
	}
	q += " ORDER BY created_at DESC, id DESC"
	if limit > 0 {
		q += " LIMIT ?"
		args = append(args, limit)
	}
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Transaction{}
	for rows.Next() {
		t, err := scanTxn(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

// UpdateTransactionStatus is a compare-and-set: it only updates the row if
// the status is still `from`, so two admins can't both apply a change.
func (s *MySQL) UpdateTransactionStatus(ctx context.Context, id int64, from, to domain.TxnStatus) error {
	res, err := s.db.ExecContext(ctx,
		"UPDATE transactions SET status = ? WHERE id = ? AND status = ?", to, id, from)
	if err != nil {
		return err
	}
	return s.casResult(ctx, res, "transactions", id)
}

// ---------- Refunds ----------

const refundCols = "id, transaction_id, amount, reason, status, requested_at, completed_at"

func scanRefund(row interface{ Scan(...any) error }) (*domain.Refund, error) {
	var r domain.Refund
	var reason string
	var requested, completed sql.NullTime
	if err := row.Scan(&r.ID, &r.TransactionID, &r.Amount, &reason, &r.Status, &requested, &completed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	r.Reason, _ = domain.ParseRefundReason(reason)
	r.RequestedAt = requested.Time
	if completed.Valid {
		r.CompletedAt = &completed.Time
	}
	return &r, nil
}

// CreateRefund relies on the UNIQUE(transaction_id) constraint, so two
// simultaneous refund requests for one payment can never both succeed.
func (s *MySQL) CreateRefund(ctx context.Context, r *domain.Refund) error {
	r.RequestedAt = time.Now().UTC().Truncate(time.Second)
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO refunds (transaction_id, amount, reason, status, requested_at) VALUES (?, ?, ?, ?, ?)",
		r.TransactionID, r.Amount, s.reason(r.Reason), r.Status, r.RequestedAt)
	if isDuplicate(err) {
		return ErrDuplicate
	}
	if err != nil {
		return err
	}
	r.ID, err = res.LastInsertId()
	return err
}

func (s *MySQL) RefundByTransaction(ctx context.Context, txnID int64) (*domain.Refund, error) {
	return scanRefund(s.db.QueryRowContext(ctx, "SELECT "+refundCols+" FROM refunds WHERE transaction_id = ?", txnID))
}

func (s *MySQL) RefundByID(ctx context.Context, id int64) (*domain.Refund, error) {
	return scanRefund(s.db.QueryRowContext(ctx, "SELECT "+refundCols+" FROM refunds WHERE id = ?", id))
}

func (s *MySQL) ListRefunds(ctx context.Context) ([]domain.Refund, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+refundCols+" FROM refunds ORDER BY requested_at DESC, id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Refund{}
	for rows.Next() {
		r, err := scanRefund(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *r)
	}
	return out, rows.Err()
}

func (s *MySQL) UpdateRefundStatus(ctx context.Context, id int64, from, to domain.RefundStatus, completedAt *time.Time) error {
	res, err := s.db.ExecContext(ctx,
		"UPDATE refunds SET status = ?, completed_at = COALESCE(?, completed_at) WHERE id = ? AND status = ?",
		to, completedAt, id, from)
	if err != nil {
		return err
	}
	return s.casResult(ctx, res, "refunds", id)
}

// ---------- Tickets ----------

const ticketCols = "id, ticket_id, user_id, transaction_id, subject, description, status, created_at, closed_at"

func scanTicket(row interface{ Scan(...any) error }) (*domain.Ticket, error) {
	var t domain.Ticket
	var created, closed sql.NullTime
	if err := row.Scan(&t.ID, &t.TicketID, &t.UserID, &t.TransactionID, &t.Subject, &t.Description, &t.Status, &created, &closed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	t.CreatedAt = created.Time
	if closed.Valid {
		t.ClosedAt = &closed.Time
	}
	return &t, nil
}

func (s *MySQL) CreateTicket(ctx context.Context, t *domain.Ticket) error {
	t.CreatedAt = time.Now().UTC().Truncate(time.Second)
	res, err := s.db.ExecContext(ctx, `INSERT INTO support_tickets
		(ticket_id, user_id, transaction_id, subject, description, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		t.TicketID, t.UserID, t.TransactionID, t.Subject, t.Description, t.Status, t.CreatedAt)
	if isDuplicate(err) {
		return ErrDuplicate
	}
	if err != nil {
		return err
	}
	t.ID, err = res.LastInsertId()
	return err
}

func (s *MySQL) TicketByPublicID(ctx context.Context, userID int64, ticketID string) (*domain.Ticket, error) {
	return scanTicket(s.db.QueryRowContext(ctx,
		"SELECT "+ticketCols+" FROM support_tickets WHERE ticket_id = ? AND user_id = ?", ticketID, userID))
}

func (s *MySQL) TicketByID(ctx context.Context, id int64) (*domain.Ticket, error) {
	return scanTicket(s.db.QueryRowContext(ctx, "SELECT "+ticketCols+" FROM support_tickets WHERE id = ?", id))
}

func (s *MySQL) ListTickets(ctx context.Context, userID int64) ([]domain.Ticket, error) {
	q := "SELECT " + ticketCols + " FROM support_tickets"
	args := []any{}
	if userID != 0 {
		q += " WHERE user_id = ?"
		args = append(args, userID)
	}
	q += " ORDER BY created_at DESC, id DESC"
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Ticket{}
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (s *MySQL) UpdateTicketStatus(ctx context.Context, id int64, from, to domain.TicketStatus, closedAt *time.Time) error {
	res, err := s.db.ExecContext(ctx,
		"UPDATE support_tickets SET status = ?, closed_at = COALESCE(?, closed_at) WHERE id = ? AND status = ?",
		to, closedAt, id, from)
	if err != nil {
		return err
	}
	return s.casResult(ctx, res, "support_tickets", id)
}

// ---------- Aggregates (computed in SQL, not by loading rows into Go) ----------

func (s *MySQL) UserSummary(ctx context.Context, userID int64) (domain.DashboardSummary, error) {
	var d domain.DashboardSummary
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
		       COALESCE(SUM(status = 'SUCCESS'), 0),
		       COALESCE(SUM(status = 'FAILED'), 0),
		       COALESCE(SUM(status = 'PENDING'), 0),
		       COALESCE(SUM(CASE WHEN status = 'SUCCESS' THEN amount END), 0)
		FROM transactions WHERE user_id = ?`, userID,
	).Scan(&d.TotalTransactions, &d.SuccessfulPayments, &d.FailedPayments, &d.PendingPayments, &d.TotalAmountPaid)
	if err != nil {
		return d, err
	}
	err = s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(r.status IN ('REQUESTED','PROCESSING')), 0),
		       COALESCE(SUM(r.status = 'COMPLETED'), 0)
		FROM refunds r JOIN transactions t ON r.transaction_id = t.id
		WHERE t.user_id = ?`, userID,
	).Scan(&d.ActiveRefunds, &d.CompletedRefunds)
	return d, err
}

func (s *MySQL) AdminSummary(ctx context.Context) (domain.AdminSummary, error) {
	var a domain.AdminSummary
	err := s.db.QueryRowContext(ctx, `
		SELECT (SELECT COUNT(*) FROM users),
		       (SELECT COUNT(*) FROM transactions),
		       (SELECT COUNT(*) FROM transactions WHERE status = 'PENDING'),
		       (SELECT COUNT(*) FROM refunds WHERE status IN ('REQUESTED','PROCESSING')),
		       (SELECT COUNT(*) FROM support_tickets WHERE status IN ('OPEN','IN_PROGRESS'))`,
	).Scan(&a.TotalUsers, &a.TotalTransactions, &a.PendingTransactions, &a.ActiveRefunds, &a.OpenTickets)
	return a, err
}
