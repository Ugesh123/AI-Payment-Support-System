package money

import (
	"encoding/json"
	"testing"
)

func TestParseAndString(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"500", "500.00", true},
		{"500.5", "500.50", true},
		{"0.10", "0.10", true},
		{"100000.00", "100000.00", true},
		{"-12.34", "-12.34", true},
		{".5", "0.50", true},
		{"1.234", "", false}, // more than 2 decimals is rejected, not rounded
		{"12.", "", false},
		{"abc", "", false},
		{"", "", false},
		{"1e5", "", false},
	}
	for _, c := range cases {
		got, err := Parse(c.in)
		if (err == nil) != c.ok {
			t.Fatalf("Parse(%q) err = %v, want ok=%v", c.in, err, c.ok)
		}
		if c.ok && got.String() != c.want {
			t.Errorf("Parse(%q) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestNoFloatingPointDrift(t *testing.T) {
	a, _ := Parse("0.10")
	b, _ := Parse("0.20")
	if (a + b).String() != "0.30" {
		t.Fatalf("0.10 + 0.20 = %s, want 0.30", a+b)
	}
}

func TestJSON(t *testing.T) {
	var v struct{ A, B Amount }
	if err := json.Unmarshal([]byte(`{"A": 250.5, "B": "99.99"}`), &v); err != nil {
		t.Fatal(err)
	}
	out, _ := json.Marshal(v)
	if string(out) != `{"A":"250.50","B":"99.99"}` {
		t.Fatalf("got %s", out)
	}
	if err := json.Unmarshal([]byte(`{"A": 1.999}`), &v); err == nil {
		t.Fatal("expected error for 3 decimal places")
	}
}

func TestScan(t *testing.T) {
	var a Amount
	for src, want := range map[any]string{"12.30": "12.30", int64(7): "7.00", nil: "0.00"} {
		var in any = src
		if s, ok := src.(string); ok {
			in = []byte(s)
		}
		if err := a.Scan(in); err != nil || a.String() != want {
			t.Errorf("Scan(%v) = %s, %v; want %s", src, a, err, want)
		}
	}
}
