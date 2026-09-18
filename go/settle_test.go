package main

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestParseAmount(t *testing.T) {
	valid := map[string]int64{
		"1250.00": 125000,
		"300.5":   30050,
		"12":      1200,
		"0.00":    0,
		"-50.00":  -5000,
		"0.01":    1,
	}
	for text, want := range valid {
		got, err := parseAmount(text)
		if err != nil {
			t.Errorf("parseAmount(%q) returned error: %v", text, err)
			continue
		}
		if got != want {
			t.Errorf("parseAmount(%q) = %d, want %d", text, got, want)
		}
	}

	for _, text := range []string{"", "1.234", "abc", "12.", "1.2.3", "10 "} {
		if _, err := parseAmount(text); !errors.Is(err, errParse) {
			t.Errorf("parseAmount(%q) should have failed with errParse, got %v", text, err)
		}
	}
}

func TestFormatAmount(t *testing.T) {
	cases := map[int64]string{
		125000: "1250.00",
		-5020:  "-50.20",
		0:      "0.00",
		7:      "0.07",
		-1:     "-0.01",
	}
	for cents, want := range cases {
		if got := formatAmount(cents); got != want {
			t.Errorf("formatAmount(%d) = %q, want %q", cents, got, want)
		}
	}
}

func TestIsValidIban(t *testing.T) {
	valid := []string{
		"XK051212012345678906",
		"XK471212000000004411",
		"DE89370400440532013000",
		"GB82WEST12345698765432",
	}
	for _, iban := range valid {
		if !isValidIban(iban) {
			t.Errorf("isValidIban(%q) = false, want true", iban)
		}
	}

	invalid := []string{
		"XK061212012345678906", // check digits altered
		"XK0512120123456789",   // too short for its scheme, fails mod-97
		"xk051212012345678906", // lower case
		"XKAB1212012345678906", // letters where check digits belong
		"",
	}
	for _, iban := range invalid {
		if isValidIban(iban) {
			t.Errorf("isValidIban(%q) = true, want false", iban)
		}
	}
}

func TestRailFees(t *testing.T) {
	cases := []struct {
		rail   Rail
		amount int64
		want   int64
	}{
		{bookRail{}, 30050, 0},
		{achRail{}, 999999, 20},
		{rtgsRail{}, 2500000, 400}, // 1.50 + 2.50
		{rtgsRail{}, 9999, 151},    // 0.009999 rounds up to a cent
		{rtgsRail{}, 1235000, 274}, // exactly .5 rounds up
		{rtgsRail{}, 1000000, 250},
	}
	for _, c := range cases {
		if got := c.rail.Fee(c.amount); got != c.want {
			t.Errorf("%s fee for %d = %d, want %d", c.rail.Name(), c.amount, got, c.want)
		}
	}
}

// TestGoldenFile runs the shared spec fixture and compares against the expected output
// that all three implementations must produce, character for character.
func TestGoldenFile(t *testing.T) {
	input, err := os.Open("../spec/payments.csv")
	if err != nil {
		t.Fatalf("cannot open fixture: %v", err)
	}
	defer input.Close()

	expected, err := os.ReadFile("../spec/expected.txt")
	if err != nil {
		t.Fatalf("cannot read expected output: %v", err)
	}

	var got strings.Builder
	if err := newProcessor().run(input, &got); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	want := strings.ReplaceAll(string(expected), "\r\n", "\n")
	if got.String() != want {
		gotLines := strings.Split(got.String(), "\n")
		wantLines := strings.Split(want, "\n")
		for i := 0; i < len(gotLines) || i < len(wantLines); i++ {
			g, w := "", ""
			if i < len(gotLines) {
				g = gotLines[i]
			}
			if i < len(wantLines) {
				w = wantLines[i]
			}
			if g != w {
				t.Fatalf("line %d:\n got: %q\nwant: %q", i+1, g, w)
			}
		}
	}
}
