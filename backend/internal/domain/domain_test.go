package domain

import "testing"

func TestRefundTransitions(t *testing.T) {
	cases := []struct {
		from, to RefundStatus
		ok       bool
	}{
		{RefundRequested, RefundProcessing, true},
		{RefundProcessing, RefundCompleted, true},
		{RefundRequested, RefundCompleted, false}, // can't skip a step
		{RefundCompleted, RefundRequested, false}, // can't go backwards
		{RefundProcessing, RefundProcessing, false},
	}
	for _, c := range cases {
		err := CheckTransition(RefundTransitions, "Refund", c.from, c.to)
		if (err == nil) != c.ok {
			t.Errorf("%s -> %s: err = %v, want ok=%v", c.from, c.to, err, c.ok)
		}
	}
}

func TestTicketTransitions(t *testing.T) {
	ok := [][2]TicketStatus{{TicketOpen, TicketInProgress}, {TicketOpen, TicketClosed}, {TicketInProgress, TicketClosed}}
	bad := [][2]TicketStatus{{TicketClosed, TicketOpen}, {TicketInProgress, TicketOpen}, {TicketClosed, TicketInProgress}}
	for _, p := range ok {
		if err := CheckTransition(TicketTransitions, "Ticket", p[0], p[1]); err != nil {
			t.Errorf("%s -> %s should be allowed: %v", p[0], p[1], err)
		}
	}
	for _, p := range bad {
		if err := CheckTransition(TicketTransitions, "Ticket", p[0], p[1]); err == nil {
			t.Errorf("%s -> %s should be rejected", p[0], p[1])
		}
	}
}

func TestTransactionFinalStates(t *testing.T) {
	if err := CheckTransition(TxnTransitions, "Transaction", TxnPending, TxnSuccess); err != nil {
		t.Fatal(err)
	}
	for _, from := range []TxnStatus{TxnSuccess, TxnFailed} {
		for _, to := range []TxnStatus{TxnSuccess, TxnFailed, TxnPending} {
			if err := CheckTransition(TxnTransitions, "Transaction", from, to); err == nil {
				t.Errorf("%s is final, but %s -> %s was allowed", from, from, to)
			}
		}
	}
}

func TestErrorMessages(t *testing.T) {
	err := CheckTransition(RefundTransitions, "Refund", RefundRequested, RefundRequested)
	if err.Error() != "Refund is already in REQUESTED status." {
		t.Errorf("got %q", err)
	}
	err = CheckTransition(RefundTransitions, "Refund", RefundRequested, RefundCompleted)
	if err.Error() != "Invalid refund status transition: REQUESTED -> COMPLETED" {
		t.Errorf("got %q", err)
	}
}

func TestEnumNamesAndValues(t *testing.T) {
	// Accepts both the JSON value and the DB name.
	for _, in := range []string{"Net Banking", "NET_BANKING"} {
		pm, ok := ParsePaymentMethod(in)
		if !ok || pm != "Net Banking" || pm.DBName() != "NET_BANKING" {
			t.Errorf("ParsePaymentMethod(%q) = %q, %v", in, pm, ok)
		}
	}
	if _, ok := ParsePaymentMethod("Cash"); ok {
		t.Error("Cash should be rejected")
	}
	r, ok := ParseRefundReason("ORDER_CANCELLED")
	if !ok || r != "Order Cancelled" {
		t.Errorf("got %q %v", r, ok)
	}
}
