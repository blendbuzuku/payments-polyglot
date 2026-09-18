#pragma once

// A flat C ABI over the C++ IBAN validator, so callers in other languages can use it.
//
// Three deliberate choices make this boundary safe to cross:
//   * extern "C"   - no C++ name mangling, so the export is called settle_iban_is_valid
//                    rather than ?isValidIban@settle@@YA_NVstring_view@std@@@Z
//   * int, not bool - C++ bool has no guaranteed size across compilers; int does
//   * const char*  - no ownership changes hands. Nothing is allocated here, so nothing
//                    can be freed by the wrong allocator, which is the usual way
//                    interop boundaries leak or crash.

#if defined(_WIN32)
#  if defined(SETTLE_IBAN_EXPORTS)
#    define SETTLE_IBAN_API __declspec(dllexport)
#  else
#    define SETTLE_IBAN_API __declspec(dllimport)
#  endif
#else
#  define SETTLE_IBAN_API __attribute__((visibility("default")))
#endif

#ifdef __cplusplus
extern "C" {
#endif

// Returns 1 when the IBAN is valid, 0 otherwise (including for a null pointer).
// The string must be null-terminated ASCII or UTF-8.
SETTLE_IBAN_API int settle_iban_is_valid(const char* iban);

#ifdef __cplusplus
}  // extern "C"
#endif
