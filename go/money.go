package main

import "fmt"

// Money is held as integer minor units (cents) in an int64. Floating point is never
// used: a binary float64 cannot represent 0.10 exactly, and a clearing file must
// balance to the cent. Core banking systems store minor units for the same reason.

// parseAmount reads "1250.00", "300.5" or "-50" as cents.
func parseAmount(text string) (int64, error) {
	i := 0
	negative := len(text) > 0 && text[0] == '-'
	if negative {
		i++
	}

	wholeStart := i
	var whole int64
	for i < len(text) && isDigit(text[i]) {
		whole = whole*10 + int64(text[i]-'0')
		i++
	}
	if i == wholeStart {
		return 0, fmt.Errorf("%w: amount has no digits: %q", errParse, text)
	}

	var fraction int64
	if i < len(text) && text[i] == '.' {
		i++
		fractionStart := i
		for i < len(text) && isDigit(text[i]) {
			fraction = fraction*10 + int64(text[i]-'0')
			i++
		}
		switch digits := i - fractionStart; digits {
		case 1:
			fraction *= 10
		case 2:
		default:
			return 0, fmt.Errorf("%w: amount needs 1 or 2 decimals: %q", errParse, text)
		}
	}

	if i != len(text) {
		return 0, fmt.Errorf("%w: unexpected characters in amount: %q", errParse, text)
	}

	cents := whole*100 + fraction
	if negative {
		return -cents, nil
	}
	return cents, nil
}

// formatAmount renders cents as "1250.00", with a leading minus only when negative.
func formatAmount(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func isUpper(c byte) bool { return c >= 'A' && c <= 'Z' }
