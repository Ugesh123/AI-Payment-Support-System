// Package api wires HTTP routes to the service layer using only the
// standard library router (Go 1.22+ method and path patterns).
package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/domain"
	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/service"
)

type ctxKey struct{}

// API holds the dependencies for every handler.
type API struct {
	svc     *service.Service
	origins map[string]bool
	log     *slog.Logger
}

// NewRouter returns the full HTTP handler with all middleware applied.
func NewRouter(svc *service.Service, allowedOrigins []string, log *slog.Logger) http.Handler {
	a := &API{svc: svc, origins: map[string]bool{}, log: log}
	for _, o := range allowedOrigins {
		if o = strings.TrimSpace(o); o != "" {
			a.origins[o] = true
		}
	}

	mux := http.NewServeMux()
	// route registers the path with and without a trailing slash, because
	// the React app calls e.g. "/transactions/" while curl users type "/transactions".
	route := func(method, path string, h http.HandlerFunc) {
		mux.HandleFunc(method+" "+path, h)
		if bare := strings.TrimSuffix(path, "/{$}"); bare != path && bare != "" {
			mux.HandleFunc(method+" "+bare, h)
		}
	}

	route("GET", "/{$}", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"message": "AI Payment Support System API (Go) is running"})
	})
	route("GET", "/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Auth
	route("POST", "/auth/register", a.register)
	route("POST", "/auth/login", a.login)
	route("GET", "/auth/me", a.user(func(w http.ResponseWriter, r *http.Request, u *domain.User) {
		writeJSON(w, http.StatusOK, u)
	}))

	// Transactions
	route("POST", "/transactions/{$}", a.user(a.createTransaction))
	route("GET", "/transactions/{$}", a.user(a.listTransactions))
	route("GET", "/transactions/{id}", a.user(a.getTransaction))

	// Refunds
	route("POST", "/refunds/{$}", a.user(a.requestRefund))
	route("GET", "/refunds/{txnID}", a.user(a.getRefund))
	route("GET", "/refunds/admin/all", a.admin(a.listAllRefunds))
	route("PATCH", "/refunds/admin/{id}/status", a.admin(a.updateRefundStatus))

	// Support tickets
	route("POST", "/tickets/{$}", a.user(a.createTicket))
	route("GET", "/tickets/{$}", a.user(a.listMyTickets))
	route("GET", "/tickets/{id}", a.user(a.getTicket))
	route("GET", "/tickets/admin/all", a.admin(a.listAllTickets))
	route("PATCH", "/tickets/admin/{id}/status", a.admin(a.updateTicketStatus))

	// Dashboard, receipt, assistant
	route("GET", "/dashboard/{$}", a.user(a.dashboard))
	route("GET", "/receipt/{txnID}", a.user(a.receipt))
	route("POST", "/assistant/chat", a.user(a.assistant))

	// Admin panel
	route("GET", "/admin/dashboard", a.admin(a.adminDashboard))
	route("GET", "/admin/users", a.admin(a.listUsers))
	route("GET", "/admin/transactions", a.admin(a.listAllTransactions))
	route("PATCH", "/admin/transactions/{id}/status", a.admin(a.updateTransactionStatus))
	route("GET", "/admin/refunds", a.admin(a.listAllRefunds))
	route("GET", "/admin/tickets", a.admin(a.listAllTickets))

	return a.recoverPanics(a.logRequests(a.cors(mux)))
}

// ---------- Helpers ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeDetail uses the error shape {"detail": "..."}, which the
// React app reads (err.response.data.detail).
func writeDetail(w http.ResponseWriter, status int, detail string) {
	writeJSON(w, status, map[string]string{"detail": detail})
}

func (a *API) writeError(w http.ResponseWriter, r *http.Request, err error) {
	var se *service.Error
	if errors.As(err, &se) {
		writeDetail(w, se.Status, se.Detail)
		return
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		writeDetail(w, http.StatusServiceUnavailable, "Request timed out")
		return
	}
	a.log.Error("internal error", "path", r.URL.Path, "err", err)
	writeDetail(w, http.StatusInternalServerError, "Internal server error")
}

// decode reads a JSON body (max 1 MB) into dst.
func decode(r *http.Request, dst any) error {
	body := http.MaxBytesReader(nil, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return &service.Error{Status: http.StatusUnprocessableEntity, Detail: "Request body is required"}
		}
		return &service.Error{Status: http.StatusUnprocessableEntity, Detail: "Invalid request body: " + err.Error()}
	}
	return nil
}

func pathID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id < 1 {
		return 0, &service.Error{Status: http.StatusUnprocessableEntity, Detail: "Invalid id: " + r.PathValue(name)}
	}
	return id, nil
}

// respond runs fn and writes either the result as JSON or the error.
func respond[T any](a *API, w http.ResponseWriter, r *http.Request, v T, err error) {
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// ---------- Middleware ----------

type userHandler func(http.ResponseWriter, *http.Request, *domain.User)

// user requires a valid "Authorization: Bearer <token>" header.
func (a *API) user(next userHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(h, "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeDetail(w, http.StatusUnauthorized, "Not authenticated")
			return
		}
		u, err := a.svc.UserFromToken(r.Context(), strings.TrimSpace(token))
		if err != nil {
			a.writeError(w, r, err)
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, u)), u)
	}
}

// admin requires a logged-in user whose role is "admin".
func (a *API) admin(next userHandler) http.HandlerFunc {
	return a.user(func(w http.ResponseWriter, r *http.Request, u *domain.User) {
		if u.Role != "admin" {
			writeDetail(w, http.StatusForbidden, "Admin access required")
			return
		}
		next(w, r, u)
	})
}

func (a *API) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" && a.origins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
			w.Header().Add("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) { s.status = code; s.ResponseWriter.WriteHeader(code) }

func (a *API) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		a.log.Info("request", "method", r.Method, "path", r.URL.Path,
			"status", rec.status, "duration_ms", time.Since(start).Milliseconds())
	})
}

func (a *API) recoverPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				a.log.Error("panic", "path", r.URL.Path, "value", v)
				writeDetail(w, http.StatusInternalServerError, "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ---------- Handlers ----------

func (a *API) register(w http.ResponseWriter, r *http.Request) {
	var in service.RegisterInput
	if err := decode(r, &in); err != nil {
		a.writeError(w, r, err)
		return
	}
	u, err := a.svc.Register(r.Context(), in)
	respond(a, w, r, u, err)
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var in service.LoginInput
	if err := decode(r, &in); err != nil {
		a.writeError(w, r, err)
		return
	}
	token, err := a.svc.Login(r.Context(), in)
	respond(a, w, r, map[string]string{"access_token": token, "token_type": "bearer"}, err)
}

func (a *API) createTransaction(w http.ResponseWriter, r *http.Request, u *domain.User) {
	var in service.TransactionInput
	if err := decode(r, &in); err != nil {
		a.writeError(w, r, err)
		return
	}
	t, err := a.svc.CreateTransaction(r.Context(), u, in)
	respond(a, w, r, t, err)
}

func (a *API) listTransactions(w http.ResponseWriter, r *http.Request, u *domain.User) {
	list, err := a.svc.ListTransactions(r.Context(), u)
	respond(a, w, r, list, err)
}

func (a *API) getTransaction(w http.ResponseWriter, r *http.Request, u *domain.User) {
	t, err := a.svc.GetTransaction(r.Context(), u, r.PathValue("id"))
	respond(a, w, r, t, err)
}

func (a *API) requestRefund(w http.ResponseWriter, r *http.Request, u *domain.User) {
	var in service.RefundInput
	if err := decode(r, &in); err != nil {
		a.writeError(w, r, err)
		return
	}
	ref, err := a.svc.RequestRefund(r.Context(), u, in)
	respond(a, w, r, ref, err)
}

func (a *API) getRefund(w http.ResponseWriter, r *http.Request, u *domain.User) {
	ref, err := a.svc.GetRefund(r.Context(), u, r.PathValue("txnID"))
	respond(a, w, r, ref, err)
}

func (a *API) listAllRefunds(w http.ResponseWriter, r *http.Request, _ *domain.User) {
	list, err := a.svc.ListAllRefunds(r.Context())
	respond(a, w, r, list, err)
}

func (a *API) updateRefundStatus(w http.ResponseWriter, r *http.Request, _ *domain.User) {
	id, err := pathID(r, "id")
	var in service.StatusInput
	if err == nil {
		err = decode(r, &in)
	}
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	ref, err := a.svc.UpdateRefundStatus(r.Context(), id, in)
	respond(a, w, r, ref, err)
}

func (a *API) createTicket(w http.ResponseWriter, r *http.Request, u *domain.User) {
	var in service.TicketInput
	if err := decode(r, &in); err != nil {
		a.writeError(w, r, err)
		return
	}
	t, err := a.svc.CreateTicket(r.Context(), u, in)
	respond(a, w, r, t, err)
}

func (a *API) listMyTickets(w http.ResponseWriter, r *http.Request, u *domain.User) {
	list, err := a.svc.ListMyTickets(r.Context(), u)
	respond(a, w, r, list, err)
}

func (a *API) getTicket(w http.ResponseWriter, r *http.Request, u *domain.User) {
	t, err := a.svc.GetTicket(r.Context(), u, r.PathValue("id"))
	respond(a, w, r, t, err)
}

func (a *API) listAllTickets(w http.ResponseWriter, r *http.Request, _ *domain.User) {
	list, err := a.svc.ListAllTickets(r.Context())
	respond(a, w, r, list, err)
}

func (a *API) updateTicketStatus(w http.ResponseWriter, r *http.Request, _ *domain.User) {
	id, err := pathID(r, "id")
	var in service.StatusInput
	if err == nil {
		err = decode(r, &in)
	}
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	t, err := a.svc.UpdateTicketStatus(r.Context(), id, in)
	respond(a, w, r, t, err)
}

func (a *API) dashboard(w http.ResponseWriter, r *http.Request, u *domain.User) {
	d, err := a.svc.Dashboard(r.Context(), u)
	respond(a, w, r, d, err)
}

func (a *API) receipt(w http.ResponseWriter, r *http.Request, u *domain.User) {
	id := r.PathValue("txnID")
	pdf, err := a.svc.Receipt(r.Context(), u, id)
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="receipt_`+id+`.pdf"`)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdf)
}

func (a *API) assistant(w http.ResponseWriter, r *http.Request, u *domain.User) {
	var in service.AssistantInput
	if err := decode(r, &in); err != nil {
		a.writeError(w, r, err)
		return
	}
	answer, err := a.svc.Assistant(r.Context(), u, in)
	respond(a, w, r, map[string]string{"answer": answer}, err)
}

func (a *API) adminDashboard(w http.ResponseWriter, r *http.Request, _ *domain.User) {
	s, err := a.svc.AdminDashboard(r.Context())
	respond(a, w, r, s, err)
}

func (a *API) listUsers(w http.ResponseWriter, r *http.Request, _ *domain.User) {
	list, err := a.svc.ListUsers(r.Context())
	respond(a, w, r, list, err)
}

func (a *API) listAllTransactions(w http.ResponseWriter, r *http.Request, _ *domain.User) {
	list, err := a.svc.ListAllTransactions(r.Context())
	respond(a, w, r, list, err)
}

func (a *API) updateTransactionStatus(w http.ResponseWriter, r *http.Request, _ *domain.User) {
	id, err := pathID(r, "id")
	var in service.StatusInput
	if err == nil {
		err = decode(r, &in)
	}
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	t, err := a.svc.UpdateTransactionStatus(r.Context(), id, in)
	respond(a, w, r, t, err)
}
