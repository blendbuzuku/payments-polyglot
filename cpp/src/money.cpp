#include "money.h"

#include <format>

#include "parse_error.h"

namespace settle {
namespace {

bool isDigit(char c) { return c >= '0' && c <= '9'; }

}  // namespace

Cents parseAmount(std::string_view text) {
    std::size_t i = 0;
    const bool negative = !text.empty() && text[0] == '-';
    if (negative) ++i;

    const std::size_t wholeStart = i;
    Cents whole = 0;
    for (; i < text.size() && isDigit(text[i]); ++i) whole = whole * 10 + (text[i] - '0');
    if (i == wholeStart) throw ParseError(std::format("amount has no digits: '{}'", text));

    Cents fraction = 0;
    if (i < text.size() && text[i] == '.') {
        const std::size_t fractionStart = ++i;
        for (; i < text.size() && isDigit(text[i]); ++i) fraction = fraction * 10 + (text[i] - '0');
        const std::size_t digits = i - fractionStart;
        if (digits == 0 || digits > 2) throw ParseError(std::format("amount needs 1 or 2 decimals: '{}'", text));
        if (digits == 1) fraction *= 10;
    }

    if (i != text.size()) throw ParseError(std::format("unexpected characters in amount: '{}'", text));

    const Cents cents = whole * 100 + fraction;
    return negative ? -cents : cents;
}

std::string formatAmount(Cents cents) {
    const Cents absolute = cents < 0 ? -cents : cents;
    return std::format("{}{}.{:02}", cents < 0 ? "-" : "", absolute / 100, absolute % 100);
}

}  // namespace settle
