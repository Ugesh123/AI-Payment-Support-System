package receipt

import (
	"bytes"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/domain"
	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/money"
)

func TestBuildProducesValidPDF(t *testing.T) {
	txn := &domain.Transaction{
		TransactionID: "TXN-1A2B3C4D", Amount: money.FromRupees(1500),
		Status: domain.TxnSuccess, PaymentMethod: "UPI", ReceiverName: "Ravi (Shop)",
		CreatedAt: time.Date(2026, 5, 1, 14, 30, 0, 0, time.UTC),
	}
	pdf := Build(txn)

	if !bytes.HasPrefix(pdf, []byte("%PDF-1.4")) || !bytes.HasSuffix(pdf, []byte("%%EOF\n")) {
		t.Fatal("missing PDF header or trailer")
	}
	for _, want := range []string{"TXN-1A2B3C4D", "Rs. 1500.00", `Ravi \(Shop\)`, "2026-05-01 02:30 PM"} {
		if !bytes.Contains(pdf, []byte(want)) {
			t.Errorf("PDF missing %q", want)
		}
	}

	// startxref must point exactly at the "xref" table.
	m := regexp.MustCompile(`startxref\n(\d+)\n`).FindSubmatch(pdf)
	if m == nil {
		t.Fatal("no startxref")
	}
	off, _ := strconv.Atoi(string(m[1]))
	if !bytes.HasPrefix(pdf[off:], []byte("xref")) {
		t.Fatalf("startxref %d does not point at xref table", off)
	}

	// Every object offset in the xref table must point at "N 0 obj".
	entries := regexp.MustCompile(`(\d{10}) 00000 n `).FindAllSubmatch(pdf, -1)
	if len(entries) != 7 {
		t.Fatalf("want 7 objects, got %d", len(entries))
	}
	for i, e := range entries {
		o, _ := strconv.Atoi(string(e[1]))
		if !bytes.HasPrefix(pdf[o:], []byte(strconv.Itoa(i+1)+" 0 obj")) {
			t.Errorf("object %d offset is wrong", i+1)
		}
	}
}
