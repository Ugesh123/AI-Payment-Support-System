// Package receipt builds a one-page PDF receipt in memory using only the
// standard library (no PDF dependency). Nothing is written to disk.
package receipt

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Ugesh123/AI-Payment-Support-System/backend/internal/domain"
)

const (
	pageW = 595.28 // A4 width in points
	pageH = 841.89 // A4 height in points
	cm    = 28.35
)

// escape makes text safe inside a PDF string literal and drops
// characters the built-in Helvetica font can't show.
func escape(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\\' || r == '(' || r == ')':
			b.WriteRune('\\')
			b.WriteRune(r)
		case r >= 32 && r < 127:
			b.WriteRune(r)
		default:
			b.WriteRune('?')
		}
	}
	return b.String()
}

// Build returns the PDF bytes for a successful transaction.
func Build(t *domain.Transaction) []byte {
	var c bytes.Buffer // page content stream
	text := func(font string, size float64, x, y float64, s string) {
		fmt.Fprintf(&c, "BT /%s %.0f Tf %.2f %.2f Td (%s) Tj ET\n", font, size, x, y, escape(s))
	}
	line := func(y float64) {
		fmt.Fprintf(&c, "%.2f %.2f m %.2f %.2f l S\n", 2*cm, y, pageW-2*cm, y)
	}

	y := pageH - 3*cm
	text("F2", 16, 2*cm, y, "AI Payment Support System")
	y -= 0.8 * cm
	text("F1", 11, 2*cm, y, "Payment Receipt")
	y -= 1.2 * cm
	line(y)
	y -= 1 * cm

	fields := [][2]string{
		{"Transaction ID", t.TransactionID},
		{"Date & Time", t.CreatedAt.Format("2006-01-02 03:04 PM")},
		{"Amount", "Rs. " + t.Amount.String()},
		{"Payment Method", string(t.PaymentMethod)},
		{"Status", string(t.Status)},
		{"Receiver Name", t.ReceiverName},
	}
	for _, f := range fields {
		text("F1", 12, 2*cm, y, f[0]+":")
		text("F1", 12, 7*cm, y, f[1])
		y -= 0.9 * cm
	}
	y -= 0.5 * cm
	line(y)
	y -= 1 * cm
	text("F3", 10, 2*cm, y, "Thank you for using our application.")

	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f] "+
			"/Resources << /Font << /F1 5 0 R /F2 6 0 R /F3 7 0 R >> >> /Contents 4 0 R >>", pageW, pageH),
		fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", c.Len(), c.String()),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Oblique >>",
	}

	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects))
	for i, obj := range objects {
		offsets[i] = out.Len()
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, off := range offsets {
		fmt.Fprintf(&out, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	return out.Bytes()
}
