package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/auth"
	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/domain"
	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/service"
	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/store"
)

// testEnv runs the real router + service on the in-memory store, with a
// payment simulator the test controls.
type testEnv struct {
	t       *testing.T
	srv     *httptest.Server
	mu      sync.Mutex
	nextPay domain.TxnStatus
}

func newEnv(t *testing.T) *testEnv {
	t.Helper()
	svc := service.New(store.NewMemory(), auth.NewTokenManager("test-secret-key-1234", time.Hour))
	e := &testEnv{t: t, nextPay: domain.TxnSuccess}
	svc.Pay = func() (domain.TxnStatus, *string) {
		e.mu.Lock()
		defer e.mu.Unlock()
		if e.nextPay == domain.TxnFailed {
			r := "Card Declined"
			return e.nextPay, &r
		}
		return e.nextPay, nil
	}
	if err := svc.EnsureAdmin(context.Background(), "admin@pay.com", "admin-pass"); err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	e.srv = httptest.NewServer(NewRouter(svc, []string{"http://localhost:5173"}, log))
	t.Cleanup(e.srv.Close)
	return e
}

func (e *testEnv) setPay(s domain.TxnStatus) { e.mu.Lock(); e.nextPay = s; e.mu.Unlock() }

// do sends a request and decodes the JSON response into out (if non-nil).
func (e *testEnv) do(method, path, token string, body any, out any) int {
	e.t.Helper()
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, e.srv.URL+path, r)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		e.t.Fatal(err)
	}
	defer res.Body.Close()
	if out != nil {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			e.t.Fatalf("%s %s: decode: %v", method, path, err)
		}
	}
	return res.StatusCode
}

type detail struct {
	Detail string `json:"detail"`
}

func (e *testEnv) signup(name, email string) string {
	e.t.Helper()
	if c := e.do("POST", "/auth/register", "", map[string]string{"name": name, "email": email, "password": "pass1234"}, nil); c != 200 {
		e.t.Fatalf("register: %d", c)
	}
	return e.loginAs(email, "pass1234")
}

func (e *testEnv) loginAs(email, pw string) string {
	e.t.Helper()
	var tok struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if c := e.do("POST", "/auth/login", "", map[string]string{"email": email, "password": pw}, &tok); c != 200 {
		e.t.Fatalf("login %s: %d", email, c)
	}
	if tok.TokenType != "bearer" || tok.AccessToken == "" {
		e.t.Fatalf("bad token response %+v", tok)
	}
	return tok.AccessToken
}

func (e *testEnv) pay(token string, amount any, method string) map[string]any {
	e.t.Helper()
	var txn map[string]any
	if c := e.do("POST", "/transactions/", token, map[string]any{"amount": amount, "receiver_name": "Swiggy", "payment_method": method}, &txn); c != 200 {
		e.t.Fatalf("create transaction: %d %v", c, txn)
	}
	return txn
}

func TestAuthFlow(t *testing.T) {
	e := newEnv(t)
	tok := e.signup("Priya", "priya@example.com")

	var me map[string]any
	if c := e.do("GET", "/auth/me", tok, nil, &me); c != 200 || me["email"] != "priya@example.com" || me["role"] != "user" {
		t.Fatalf("me: %d %v", c, me)
	}
	if _, leaked := me["password_hash"]; leaked {
		t.Fatal("password hash must never be returned")
	}

	var d detail
	if c := e.do("POST", "/auth/register", "", map[string]string{"name": "P", "email": "priya@example.com", "password": "x"}, &d); c != 400 || d.Detail != "Email already registered" {
		t.Fatalf("duplicate register: %d %q", c, d.Detail)
	}
	if c := e.do("POST", "/auth/register", "", map[string]string{"name": "P", "email": "not-an-email", "password": "x"}, nil); c != 422 {
		t.Fatalf("bad email: %d", c)
	}
	if c := e.do("POST", "/auth/login", "", map[string]string{"email": "priya@example.com", "password": "wrong"}, &d); c != 401 || d.Detail != "Invalid email or password" {
		t.Fatalf("wrong password: %d %q", c, d.Detail)
	}
	if c := e.do("POST", "/auth/login", "", map[string]string{"email": "nobody@example.com", "password": "x"}, &d); c != 401 || d.Detail != "Invalid email or password" {
		t.Fatalf("unknown user should get the same message: %d %q", c, d.Detail)
	}
	if c := e.do("GET", "/auth/me", "", nil, nil); c != 401 {
		t.Fatalf("no token: %d", c)
	}
	if c := e.do("GET", "/auth/me", "garbage.token.here", nil, nil); c != 401 {
		t.Fatalf("bad token: %d", c)
	}
}

func TestTransactionValidationAndOwnership(t *testing.T) {
	e := newEnv(t)
	alice := e.signup("Alice", "alice@example.com")
	bob := e.signup("Bob", "bob@example.com")

	bad := []map[string]any{
		{"amount": 0, "receiver_name": "X", "payment_method": "UPI"},
		{"amount": 0.5, "receiver_name": "X", "payment_method": "UPI"},
		{"amount": 100000.01, "receiver_name": "X", "payment_method": "UPI"},
		{"amount": 10.555, "receiver_name": "X", "payment_method": "UPI"},
		{"amount": 10, "receiver_name": "", "payment_method": "UPI"},
		{"amount": 10, "receiver_name": "X", "payment_method": "Cash"},
		{"receiver_name": "X", "payment_method": "UPI"},
	}
	for _, b := range bad {
		if c := e.do("POST", "/transactions/", alice, b, nil); c != 422 {
			t.Errorf("%v: want 422, got %d", b, c)
		}
	}

	txn := e.pay(alice, "1499.50", "Net Banking")
	if txn["amount"] != "1499.50" || txn["payment_method"] != "Net Banking" || txn["status"] != "SUCCESS" {
		t.Fatalf("unexpected transaction %v", txn)
	}
	id := txn["transaction_id"].(string)
	if !strings.HasPrefix(id, "TXN-") || len(id) != 12 {
		t.Fatalf("bad id %q", id)
	}

	var list []map[string]any
	if c := e.do("GET", "/transactions", alice, nil, &list); c != 200 || len(list) != 1 { // no trailing slash
		t.Fatalf("list: %d %d", c, len(list))
	}
	if c := e.do("GET", "/transactions/"+id, alice, nil, nil); c != 200 {
		t.Fatalf("get own: %d", c)
	}
	if c := e.do("GET", "/transactions/"+id, bob, nil, nil); c != 404 {
		t.Fatalf("bob must not see alice's transaction: %d", c)
	}
}

func TestRefundRules(t *testing.T) {
	e := newEnv(t)
	tok := e.signup("Alice", "alice@example.com")
	admin := e.loginAs("admin@pay.com", "admin-pass")

	e.setPay(domain.TxnFailed)
	failed := e.pay(tok, 200, "Card")
	if failed["failure_reason"] != "Card Declined" {
		t.Fatalf("failure_reason missing: %v", failed)
	}
	var d detail
	if c := e.do("POST", "/refunds/", tok, map[string]string{"transaction_id": failed["transaction_id"].(string), "refund_reason": "Other"}, &d); c != 400 {
		t.Fatalf("refund on failed payment: %d %q", c, d.Detail)
	}

	e.setPay(domain.TxnSuccess)
	ok := e.pay(tok, 999, "UPI")
	txnID := ok["transaction_id"].(string)
	if c := e.do("POST", "/refunds/", tok, map[string]string{"transaction_id": txnID, "refund_reason": "Refund please"}, nil); c != 422 {
		t.Fatalf("invalid reason: %d", c)
	}

	var ref map[string]any
	if c := e.do("POST", "/refunds/", tok, map[string]string{"transaction_id": txnID, "refund_reason": "Wrong Amount"}, &ref); c != 200 {
		t.Fatalf("refund: %d", c)
	}
	if ref["amount"] != "999.00" || ref["status"] != "REQUESTED" || ref["reason"] != "Wrong Amount" || ref["completed_at"] != nil {
		t.Fatalf("unexpected refund %v", ref)
	}
	if c := e.do("POST", "/refunds/", tok, map[string]string{"transaction_id": txnID, "refund_reason": "Other"}, &d); c != 400 || !strings.Contains(d.Detail, "already been requested") {
		t.Fatalf("second refund: %d %q", c, d.Detail)
	}
	if c := e.do("GET", "/refunds/"+txnID, tok, nil, nil); c != 200 {
		t.Fatalf("refund status: %d", c)
	}

	// Admin state machine: REQUESTED -> COMPLETED is not allowed (must go through PROCESSING).
	path := "/refunds/admin/" + jsonID(ref) + "/status"
	if c := e.do("PATCH", path, tok, map[string]string{"status": "PROCESSING"}, nil); c != 403 {
		t.Fatalf("normal user calling admin route: %d", c)
	}
	if c := e.do("PATCH", path, admin, map[string]string{"status": "COMPLETED"}, &d); c != 400 || d.Detail != "Invalid refund status transition: REQUESTED -> COMPLETED" {
		t.Fatalf("skip step: %d %q", c, d.Detail)
	}
	if c := e.do("PATCH", path, admin, map[string]string{"status": "PROCESSING"}, &ref); c != 200 || ref["status"] != "PROCESSING" {
		t.Fatalf("to processing: %d %v", c, ref)
	}
	if c := e.do("PATCH", path, admin, map[string]string{"status": "PROCESSING"}, &d); c != 400 || d.Detail != "Refund is already in PROCESSING status." {
		t.Fatalf("same status: %d %q", c, d.Detail)
	}
	if c := e.do("PATCH", path, admin, map[string]string{"status": "COMPLETED"}, &ref); c != 200 || ref["completed_at"] == nil {
		t.Fatalf("to completed: %d %v", c, ref)
	}
	if c := e.do("PATCH", "/refunds/admin/9999/status", admin, map[string]string{"status": "PROCESSING"}, nil); c != 404 {
		t.Fatalf("missing refund: %d", c)
	}
}

// Ten requests race to refund the same payment; exactly one must succeed.
func TestConcurrentRefundRequests(t *testing.T) {
	e := newEnv(t)
	tok := e.signup("Alice", "alice@example.com")
	txnID := e.pay(tok, 500, "Wallet")["transaction_id"].(string)

	codes := make(chan int, 10)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			codes <- e.do("POST", "/refunds/", tok, map[string]string{"transaction_id": txnID, "refund_reason": "Duplicate Payment"}, nil)
		}()
	}
	wg.Wait()
	close(codes)
	okCount := 0
	for c := range codes {
		if c == 200 {
			okCount++
		} else if c != 400 {
			t.Errorf("unexpected status %d", c)
		}
	}
	if okCount != 1 {
		t.Fatalf("%d refunds created for one payment, want exactly 1", okCount)
	}
}

func TestTicketsAndAdminTransactionUpdate(t *testing.T) {
	e := newEnv(t)
	tok := e.signup("Alice", "alice@example.com")
	admin := e.loginAs("admin@pay.com", "admin-pass")

	e.setPay(domain.TxnPending)
	txn := e.pay(tok, 300, "UPI")
	txnID := txn["transaction_id"].(string)

	var tk map[string]any
	if c := e.do("POST", "/tickets/", tok, map[string]string{"transaction_id": txnID, "subject": "Stuck payment", "description": "Money debited but pending"}, &tk); c != 200 {
		t.Fatalf("ticket: %d", c)
	}
	if !strings.HasPrefix(tk["ticket_id"].(string), "TKT-") || tk["status"] != "OPEN" {
		t.Fatalf("unexpected ticket %v", tk)
	}
	if c := e.do("POST", "/tickets/", tok, map[string]string{"transaction_id": txnID, "subject": "", "description": "x"}, nil); c != 422 {
		t.Fatalf("empty subject: %d", c)
	}
	var mine []map[string]any
	if c := e.do("GET", "/tickets/", tok, nil, &mine); c != 200 || len(mine) != 1 {
		t.Fatalf("my tickets: %d %d", c, len(mine))
	}
	if c := e.do("GET", "/tickets/"+tk["ticket_id"].(string), tok, nil, nil); c != 200 {
		t.Fatalf("get ticket: %d", c)
	}

	tpath := "/tickets/admin/" + jsonID(tk) + "/status"
	if c := e.do("PATCH", tpath, admin, map[string]string{"status": "CLOSED"}, &tk); c != 200 || tk["closed_at"] == nil {
		t.Fatalf("close ticket: %d %v", c, tk)
	}
	if c := e.do("PATCH", tpath, admin, map[string]string{"status": "OPEN"}, nil); c != 400 {
		t.Fatalf("reopen closed ticket: %d", c)
	}

	// PENDING -> SUCCESS is allowed once; SUCCESS is final.
	path := "/admin/transactions/" + jsonID(txn) + "/status"
	if c := e.do("PATCH", path, admin, map[string]string{"status": "SUCCESS"}, &txn); c != 200 || txn["status"] != "SUCCESS" {
		t.Fatalf("pending->success: %d %v", c, txn)
	}
	if c := e.do("PATCH", path, admin, map[string]string{"status": "FAILED"}, nil); c != 400 {
		t.Fatalf("success is final: %d", c)
	}
	if c := e.do("PATCH", "/admin/transactions/abc/status", admin, map[string]string{"status": "FAILED"}, nil); c != 422 {
		t.Fatalf("non-numeric id: %d", c)
	}
}

func TestDashboards(t *testing.T) {
	e := newEnv(t)
	tok := e.signup("Alice", "alice@example.com")
	admin := e.loginAs("admin@pay.com", "admin-pass")

	e.pay(tok, "100.10", "UPI")
	e.pay(tok, "200.20", "Card")
	e.setPay(domain.TxnFailed)
	e.pay(tok, 50, "UPI")
	e.setPay(domain.TxnPending)
	e.pay(tok, 75, "UPI")

	var d struct {
		Summary            map[string]any   `json:"summary"`
		RecentTransactions []map[string]any `json:"recent_transactions"`
	}
	if c := e.do("GET", "/dashboard/", tok, nil, &d); c != 200 {
		t.Fatalf("dashboard: %d", c)
	}
	s := d.Summary
	if s["total_transactions"] != 4.0 || s["successful_payments"] != 2.0 || s["failed_payments"] != 1.0 ||
		s["pending_payments"] != 1.0 || s["total_amount_paid"] != "300.30" {
		t.Fatalf("summary wrong: %v", s)
	}
	if len(d.RecentTransactions) != 4 || d.RecentTransactions[0]["amount"] != "75.00" {
		t.Fatalf("recent should be newest first: %v", d.RecentTransactions)
	}

	var a map[string]float64
	if c := e.do("GET", "/admin/dashboard", admin, nil, &a); c != 200 {
		t.Fatalf("admin dashboard: %d", c)
	}
	if a["total_users"] != 2 || a["total_transactions"] != 4 || a["pending_transactions"] != 1 {
		t.Fatalf("admin summary wrong: %v", a)
	}
	if c := e.do("GET", "/admin/users", tok, nil, nil); c != 403 {
		t.Fatalf("user on admin route: %d", c)
	}
	for _, p := range []string{"/admin/users", "/admin/transactions", "/admin/refunds", "/admin/tickets", "/refunds/admin/all", "/tickets/admin/all"} {
		var list []any
		if c := e.do("GET", p, admin, nil, &list); c != 200 {
			t.Errorf("%s: %d", p, c)
		}
	}
}

func TestReceiptAndAssistant(t *testing.T) {
	e := newEnv(t)
	tok := e.signup("Alice", "alice@example.com")
	id := e.pay(tok, 1200, "UPI")["transaction_id"].(string)

	req, _ := http.NewRequest("GET", e.srv.URL+"/receipt/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != 200 || res.Header.Get("Content-Type") != "application/pdf" ||
		!strings.Contains(res.Header.Get("Content-Disposition"), "receipt_"+id+".pdf") || !bytes.HasPrefix(body, []byte("%PDF")) {
		t.Fatalf("receipt: %d %v", res.StatusCode, res.Header)
	}

	e.setPay(domain.TxnFailed)
	failedID := e.pay(tok, 10, "UPI")["transaction_id"].(string)
	if c := e.do("GET", "/receipt/"+failedID, tok, nil, nil); c != 400 {
		t.Fatalf("receipt for failed payment: %d", c)
	}

	var ans map[string]string
	if c := e.do("POST", "/assistant/chat", tok, map[string]string{"transaction_id": id, "question": "where is my money?"}, &ans); c != 200 || !strings.Contains(ans["answer"], id) {
		t.Fatalf("assistant: %d %v", c, ans)
	}
}

func TestCORS(t *testing.T) {
	e := newEnv(t)
	req, _ := http.NewRequest("OPTIONS", e.srv.URL+"/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 204 || res.Header.Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("preflight: %d %v", res.StatusCode, res.Header)
	}

	req, _ = http.NewRequest("OPTIONS", e.srv.URL+"/auth/login", nil)
	req.Header.Set("Origin", "https://evil.example")
	res, _ = http.DefaultClient.Do(req)
	res.Body.Close()
	if res.Header.Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("unknown origin must not be allowed")
	}
}

func jsonID(m map[string]any) string {
	b, _ := json.Marshal(m["id"])
	return string(b)
}
