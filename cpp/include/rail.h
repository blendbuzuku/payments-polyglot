#pragma once

#include <string_view>

#include "money.h"

namespace settle {

// A clearing rail: what a payment travels on, and what that costs.
class Rail {
public:
    virtual ~Rail() = default;  // deleted through a Rail*, so this must be virtual

    virtual std::string_view name() const = 0;
    virtual Cents fee(Cents amount) const = 0;
};

// Both parties bank with us, so the transfer never leaves the books. Free.
class BookRail final : public Rail {
public:
    std::string_view name() const override { return "BOOK"; }
    Cents fee(Cents) const override { return 0; }
};

// Batch clearing: cheap, settled in the next cycle.
class AchRail final : public Rail {
public:
    std::string_view name() const override { return "ACH"; }
    Cents fee(Cents) const override { return 20; }
};

// Real-time gross settlement: 1.50 plus 0.01% of the amount, rounded half-up.
class RtgsRail final : public Rail {
public:
    std::string_view name() const override { return "RTGS"; }

    // 0.01% of `amount` is amount/10000 cents; adding 5000 before the integer
    // division rounds half-up without ever touching a double.
    Cents fee(Cents amount) const override { return 150 + (amount + 5000) / 10000; }
};

}  // namespace settle
