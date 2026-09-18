package main

// isValidIban checks an IBAN per ISO 13616: structure plus the ISO 7064 MOD-97-10 check.
func isValidIban(iban string) bool {
	if len(iban) < 15 || len(iban) > 34 {
		return false
	}

	for i := 0; i < len(iban); i++ {
		c := iban[i]
		var ok bool
		switch {
		case i < 2:
			ok = isUpper(c) // country code
		case i < 4:
			ok = isDigit(c) // check digits
		default:
			ok = isUpper(c) || isDigit(c)
		}
		if !ok {
			return false
		}
	}

	// Move the first four characters to the end, map A=10..Z=35, then take mod 97.
	// The rearranged number is far too large for any integer type, so the remainder
	// is folded in digit by digit.
	remainder := 0
	for k := 0; k < len(iban); k++ {
		c := iban[(k+4)%len(iban)]
		if isDigit(c) {
			remainder = (remainder*10 + int(c-'0')) % 97
		} else {
			remainder = (remainder*100 + int(c-'A') + 10) % 97
		}
	}

	return remainder == 1
}

// bankCode returns the four-digit bank code: characters 5-8 of the IBAN.
func bankCode(iban string) string { return iban[4:8] }
