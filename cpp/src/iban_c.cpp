#include "iban_c.h"

#include "iban.h"

// The C entry point is a thin shell over the C++ implementation: it checks the
// pointer, converts to the C++ type, and maps the result onto an int. No exception
// may escape here, because unwinding across a C ABI boundary is undefined behaviour.
// settle::isValidIban does not throw, so there is nothing to catch.
int settle_iban_is_valid(const char* iban) {
    if (iban == nullptr) return 0;
    return settle::isValidIban(iban) ? 1 : 0;
}
