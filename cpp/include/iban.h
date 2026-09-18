#pragma once

#include <string_view>

namespace settle {

// IBAN validation per ISO 13616: structure plus the ISO 7064 MOD-97-10 check.
bool isValidIban(std::string_view iban);

// The four-digit bank code: characters 5-8 of the IBAN.
std::string_view bankCode(std::string_view iban);

}  // namespace settle
