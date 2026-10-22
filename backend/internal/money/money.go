// Package money stores amounts as whole paise (int64) instead of floats,
// so ₹0.10 + ₹0.20 is always exactly ₹0.30. MySQL stores it as NUMERIC(10,2).
package money

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Amount is a money value in paise (1 rupee = 100 paise).
type Amount int64

var ErrInvalidAmount = errors.New("invalid amount: use a number with at most 2 decimal places")

// FromRupees builds an Amount from whole rupees.
func FromRupees(r int64) Amount { return Amount(r * 100) }

// Parse turns "500", "500.5" or "500.50" into paise.
// More than 2 decimal places is rejected rather than silently rounded.
func Parse(s string) (Amount, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrInvalidAmount
	}
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(strings.TrimPrefix(s, "-"), "+")

	whole, frac, hasDot := strings.Cut(s, ".")
	if whole == "" {
		whole = "0"
	}
	if hasDot && (frac == "" || len(frac) > 2) {
		return 0, ErrInvalidAmount
	}
	for len(frac) < 2 {
		frac += "0"
	}
	if !digitsOnly(whole) || !digitsOnly(frac) || len(whole) > 15 {
		return 0, ErrInvalidAmount
	}
	w, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, ErrInvalidAmount
	}
	f, _ := strconv.ParseInt(frac, 10, 64)
	v := w*100 + f
	if neg {
		v = -v
	}
	return Amount(v), nil
}

func digitsOnly(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// String formats paise as "500.00".
func (a Amount) String() string {
	v := int64(a)
	sign := ""
	if v < 0 {
		sign, v = "-", -v
	}
	return fmt.Sprintf("%s%d.%02d", sign, v/100, v%100)
}

// MarshalJSON writes the amount as a string ("500.00") so the frontend
// never sees float rounding.
func (a Amount) MarshalJSON() ([]byte, error) { return json.Marshal(a.String()) }

// UnmarshalJSON accepts both 500.5 and "500.50".
func (a *Amount) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	v, err := Parse(s)
	if err != nil {
		return err
	}
	*a = v
	return nil
}

// Value stores the amount in a DECIMAL(10,2) column.
func (a Amount) Value() (driver.Value, error) { return a.String(), nil }

// Scan reads a DECIMAL column (the MySQL driver returns it as []byte).
func (a *Amount) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*a = 0
		return nil
	case []byte:
		p, err := Parse(string(v))
		*a = p
		return err
	case string:
		p, err := Parse(v)
		*a = p
		return err
	case int64:
		*a = Amount(v * 100)
		return nil
	case float64:
		p, err := Parse(strconv.FormatFloat(v, 'f', 2, 64))
		*a = p
		return err
	}
	return fmt.Errorf("money: cannot scan %T", src)
}
