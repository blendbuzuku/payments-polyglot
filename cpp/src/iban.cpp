#include "iban.h"

#include <cstddef>

namespace settle {
namespace {

bool isDigit(char c) { return c >= '0' && c <= '9'; }
bool isUpper(char c) { return c >= 'A' && c <= 'Z'; }

}  // namespace

bool isValidIban(std::string_view iban) {
    if (iban.size() < 15 || iban.size() > 34) return false;

    for (std::size_t i = 0; i < iban.size(); ++i) {
        const char c = iban[i];
        const bool ok = i < 2   ? isUpper(c)                 // country code
                        : i < 4 ? isDigit(c)                 // check digits
                                : isUpper(c) || isDigit(c);
        if (!ok) return false;
    }

    // Move the first four characters to the end, map A=10..Z=35, then take mod 97.
    // The rearranged number is far too large for any integer type, so the remainder
    // is folded in digit by digit.
    int remainder = 0;
    for (std::size_t k = 0; k < iban.size(); ++k) {
        const char c = iban[(k + 4) % iban.size()];
        remainder = isDigit(c) ? (remainder * 10 + (c - '0')) % 97
                               : (remainder * 100 + (c - 'A' + 10)) % 97;
    }

    return remainder == 1;
}

std::string_view bankCode(std::string_view iban) { return iban.substr(4, 4); }

}  // namespace settle
