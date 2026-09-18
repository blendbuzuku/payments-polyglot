package main

// Rail is a clearing rail: what a payment travels on, and what that costs.
// Go has no classes; an interface plus three small types does the same job.
type Rail interface {
	Name() string
	Fee(amount int64) int64
}

// bookRail: both parties bank with us, so the transfer never leaves the books. Free.
type bookRail struct{}

func (bookRail) Name() string    { return "BOOK" }
func (bookRail) Fee(int64) int64 { return 0 }

// achRail: batch clearing, cheap, settled in the next cycle.
type achRail struct{}

func (achRail) Name() string    { return "ACH" }
func (achRail) Fee(int64) int64 { return 20 }

// rtgsRail: real-time gross settlement, 1.50 plus 0.01% of the amount, rounded half-up.
type rtgsRail struct{}

func (rtgsRail) Name() string { return "RTGS" }

// 0.01% of amount is amount/10000 cents; adding 5000 before the integer division
// rounds half-up without ever touching a float.
func (rtgsRail) Fee(amount int64) int64 { return 150 + (amount+5000)/10000 }
