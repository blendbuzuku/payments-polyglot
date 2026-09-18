#pragma once

#include <cstdint>
#include <string>
#include <string_view>

namespace settle {

// Money is held as integer minor units (cents). Floating point is never used:
// a binary double cannot represent 0.10 exactly, and a clearing file must balance
// to the cent. Core banking systems store minor units for the same reason.
using Cents = std::int64_t;

// Parses "1250.00", "300.5" or "-50" into cents. Throws ParseError otherwise.
Cents parseAmount(std::string_view text);

// Formats cents as "1250.00", with a leading minus only when negative.
std::string formatAmount(Cents cents);

}  // namespace settle
