package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

const (
	supportedCurrency = "EUR"
	rtgsThreshold     = 1_000_000 // 10,000.00
	feeAccount        = "FEES"
)

// errParse marks a line that cannot be read as a payment at all, as opposed to one
// that fails a business rule. Go has no exceptions: errors are values, and callers
// match on them with errors.Is.
var errParse = errors.New("parse error")

type payment struct {
	id       string
	debtor   string
	creditor string
	amount   int64
	currency string
	urgent   bool
}

type railTotals struct {
	count  int
	volume int64
	fees   int64
}

// processor reads a clearing file, applies the scheme rules, routes each accepted
// payment and prints the report. Rejection reasons are checked in a fixed order and
// the first failure wins, so a line always has exactly one reason.
type processor struct {
	rails     []Rail
	totals    map[string]*railTotals
	seenIDs   map[string]bool
	positions map[string]int64
	accepted  int
	rejected  int
}

func newProcessor() *processor {
	p := &processor{
		rails:     []Rail{bookRail{}, achRail{}, rtgsRail{}},
		totals:    make(map[string]*railTotals),
		seenIDs:   make(map[string]bool),
		positions: make(map[string]int64),
	}
	for _, rail := range p.rails {
		p.totals[rail.Name()] = &railTotals{}
	}
	return p
}

func (p *processor) run(input io.Reader, output io.Writer) error {
	if _, err := fmt.Fprintln(output, "== Payments =="); err != nil {
		return err
	}

	scanner := bufio.NewScanner(input)
	header := true
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r") // a file written on Windows
		if header {
			header = false
			continue
		}
		if line == "" {
			continue
		}
		p.process(line, output)
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	fmt.Fprintln(output, "== Rails ==")
	for _, rail := range p.rails {
		t := p.totals[rail.Name()]
		fmt.Fprintf(output, "%s count=%d volume=%s fees=%s\n",
			rail.Name(), t.count, formatAmount(t.volume), formatAmount(t.fees))
	}

	fmt.Fprintln(output, "== Positions ==")
	accounts := make([]string, 0, len(p.positions))
	for account := range p.positions {
		accounts = append(accounts, account)
	}
	sort.Strings(accounts) // map iteration order is random in Go, so sort explicitly

	var sum int64
	for _, account := range accounts {
		fmt.Fprintf(output, "%s %s\n", account, formatAmount(p.positions[account]))
		sum += p.positions[account]
	}

	status := "OK"
	if sum != 0 {
		status = "MISMATCH"
	}
	fmt.Fprintln(output, "== Check ==")
	_, err := fmt.Fprintf(output, "accepted=%d rejected=%d sum=%s %s\n",
		p.accepted, p.rejected, formatAmount(sum), status)
	return err
}

func (p *processor) process(line string, output io.Writer) {
	fields := strings.Split(line, ",")

	pmt, err := readPayment(fields)
	if err != nil {
		id := ""
		if len(fields) > 0 {
			id = fields[0]
		}
		p.reject(id, "MALFORMED", output)
		return
	}

	if reason := p.check(pmt); reason != "" {
		p.reject(pmt.id, reason, output)
		return
	}

	rail := p.route(pmt)
	fee := rail.Fee(pmt.amount)
	p.post(pmt, fee)

	t := p.totals[rail.Name()]
	t.count++
	t.volume += pmt.amount
	t.fees += fee

	p.accepted++
	fmt.Fprintf(output, "%s ACCEPTED %s %s fee=%s\n",
		pmt.id, rail.Name(), formatAmount(pmt.amount), formatAmount(fee))
}

func readPayment(fields []string) (payment, error) {
	if len(fields) != 6 {
		return payment{}, fmt.Errorf("%w: expected 6 fields, got %d", errParse, len(fields))
	}

	amount, err := parseAmount(fields[3])
	if err != nil {
		return payment{}, err
	}

	urgent, err := parseFlag(fields[5])
	if err != nil {
		return payment{}, err
	}

	return payment{
		id:       fields[0],
		debtor:   fields[1],
		creditor: fields[2],
		amount:   amount,
		currency: fields[4],
		urgent:   urgent,
	}, nil
}

func parseFlag(text string) (bool, error) {
	switch text {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("%w: urgent must be true or false: %q", errParse, text)
	}
}

// check returns the first rule that fails, or "" when the payment is good.
func (p *processor) check(pmt payment) string {
	if p.seenIDs[pmt.id] {
		return "DUPLICATE_ID"
	}
	p.seenIDs[pmt.id] = true

	if !isValidIban(pmt.debtor) || !isValidIban(pmt.creditor) {
		return "INVALID_IBAN"
	}
	if pmt.debtor == pmt.creditor {
		return "SAME_ACCOUNT"
	}
	if pmt.amount <= 0 {
		return "INVALID_AMOUNT"
	}
	if pmt.currency != supportedCurrency {
		return "UNSUPPORTED_CURRENCY"
	}
	return ""
}

func (p *processor) route(pmt payment) Rail {
	// Same bank: the money never leaves our books, whatever the sender asked for.
	if bankCode(pmt.debtor) == bankCode(pmt.creditor) {
		return p.rails[0]
	}
	if pmt.urgent || pmt.amount >= rtgsThreshold {
		return p.rails[2]
	}
	return p.rails[1]
}

func (p *processor) reject(id, reason string, output io.Writer) {
	fmt.Fprintf(output, "%s REJECTED %s\n", id, reason)
	p.rejected++
}

// post writes the double entry: the sender pays amount plus fee, the beneficiary
// receives the amount, and FEES takes the difference, so the batch nets to zero.
func (p *processor) post(pmt payment, fee int64) {
	p.positions[pmt.debtor] -= pmt.amount + fee
	p.positions[pmt.creditor] += pmt.amount
	p.positions[feeAccount] += fee
}
