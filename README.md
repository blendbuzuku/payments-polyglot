# payments-polyglot

One payment batch processor, written three times: **C# (.NET 8)**, **C++20** and **Go**.

Same input file, byte-identical output, three sets of language idioms. It exists because
"I can pick up your stack" is easy to claim in a cover letter and hard to prove, and because
the interesting parts of payments code — money that must not lose a cent, rules that must be
applied in a fixed order, a ledger that must net to zero — look different in a language with
exceptions and a GC, a language with destructors and manual ownership, and a language with
neither.

## What it does

`settle` reads a clearing file in the shape a domestic payment scheme would send, validates
each instruction, routes the accepted ones onto a rail, and prints the resulting book.

```
$ settle spec/payments.csv
== Payments ==
P001 ACCEPTED ACH 1250.00 fee=0.20
P002 ACCEPTED BOOK 300.50 fee=0.00
P003 ACCEPTED RTGS 25000.00 fee=4.00
P005 REJECTED INVALID_IBAN
...
== Rails ==
BOOK count=2 volume=720.50 fees=0.00
ACH count=2 volume=11249.99 fees=0.40
RTGS count=4 volume=47449.99 fees=10.75
== Positions ==
FEES 11.15
XK051212012345678906 -1450.71
...
== Check ==
accepted=8 rejected=9 sum=0.00 OK
```

The last line is the point. Every accepted payment debits the sender for amount plus fee,
credits the beneficiary with the amount, and credits a `FEES` account with the fee, so the
positions must sum to exactly `0.00`. It is the same end-of-day check a real settlement
system runs, and the reason the whole program stores money as integer cents rather than
floating point.

## Rules

**Validation** — the first rule that fails wins, so every rejected line has exactly one reason:

| Reason | Meaning |
| --- | --- |
| `MALFORMED` | Not 6 fields, an unparseable amount, or `urgent` that isn't `true`/`false` |
| `DUPLICATE_ID` | The id already appeared on an earlier readable line |
| `INVALID_IBAN` | Debtor or creditor fails ISO 13616 structure or the MOD-97-10 check |
| `SAME_ACCOUNT` | Debtor equals creditor |
| `INVALID_AMOUNT` | Amount is zero or negative |
| `UNSUPPORTED_CURRENCY` | Currency is not `EUR` |

**Routing** — the first rail that applies:

| Rail | When | Fee |
| --- | --- | --- |
| `BOOK` | Debtor and creditor share a bank code (IBAN characters 5–8) | 0.00 |
| `RTGS` | Urgent, or 10,000.00 and above | 1.50 + 0.01% of the amount, rounded half-up |
| `ACH` | Everything else | 0.20 |

The IBAN check folds the MOD-97 remainder in digit by digit, because the rearranged IBAN is a
number far too large for any 64-bit integer. The RTGS percentage rounds half-up in integer
arithmetic (`150 + (amount + 5000) / 10000`), never through a float.

## Running it

```bash
# C# (.NET 8)
cd csharp/Settle && dotnet run -- ../../spec/payments.csv

# C++20 on Windows: finds Visual Studio itself, no Developer prompt needed
cpp\build.bat && cpp\build\settle.exe spec\payments.csv

# C++20 anywhere else
cd cpp && cmake -B build && cmake --build build && ./build/settle ../spec/payments.csv
g++ -std=c++20 -Wall -Wextra -Iinclude src/*.cpp -o settle                # GCC

# Go
cd go && go run . ../spec/payments.csv
go test ./...
```

Every implementation must reproduce `spec/expected.txt` exactly. CI builds all three on every
push and diffs their output against that file, so a change that breaks one language fails the
build.

## What each language made me do differently

| | C# | C++ | Go |
| --- | --- | --- | --- |
| Rail abstraction | `abstract class` + `sealed` overrides | abstract base with a **virtual destructor**, held in `unique_ptr` | small `interface`, values with no inheritance at all |
| Failed parse | `throw ParseException` | `throw ParseError`, caught by `const&` | `error` value wrapping `errParse`, checked with `errors.Is` |
| Ordered report | `SortedDictionary` with `StringComparer.Ordinal` | `std::map`, ordered by definition | `map` + explicit `sort.Strings`, because Go randomises map order |
| Cleanup | `using` on the reader | destructors, on every path out of the scope | `defer file.Close()` |
| Money | `long` cents | `std::int64_t` cents | `int64` cents |

The C# version needs `StringComparer.Ordinal` explicitly: the default comparer is
culture-sensitive, which would sort the accounts differently from the other two on some
machines. That class of bug is exactly what writing the same program three times exposes.

## Interop: C# calling the C++ validator

The IBAN check also builds as a shared library with a flat C ABI, and the .NET app can
use it instead of its own implementation:

```bash
# Linux
g++ -std=c++20 -O2 -fPIC -shared -Icpp/include cpp/src/iban.cpp cpp/src/iban_c.cpp -o libsettle_iban.so
dotnet publish csharp/Settle/Settle.csproj -c Release -o publish && cp libsettle_iban.so publish/
./publish/settle spec/payments.csv --native

# Windows (Developer PowerShell)
cl /LD /std:c++20 /EHsc /W4 /DSETTLE_IBAN_EXPORTS /Iinclude cpp\src\iban.cpp cpp\src\iban_c.cpp /Fe:settle_iban.dll
```

`--native` routes every IBAN through `settle_iban_is_valid` in the C++ library; without it
the managed implementation answers. Both must print the same report, and CI asserts exactly
that on every push.

Four things decide whether a boundary like this works, and all four are visible in
[`iban_c.h`](cpp/include/iban_c.h) and [`NativeIbanValidator.cs`](csharp/Settle/Interop/NativeIbanValidator.cs):

| Concern | Choice here | What goes wrong otherwise |
| --- | --- | --- |
| Symbol name | `extern "C"` | The export is mangled (`?isValidIban@settle@@…`) and .NET throws `EntryPointNotFoundException` |
| Types | `int`, not `bool` | C++ `bool` has no guaranteed size across compilers |
| Strings | `LPUTF8Str` marshalling onto `const char*` | .NET's default is UTF-16, which a `char*` is not |
| Memory | nothing is allocated across the boundary | Freeing with the wrong allocator corrupts the heap |

Bitness matters too: a 64-bit process cannot load a 32-bit library, which is what
`BadImageFormatException` means when it appears.

The Go implementation deliberately does not do this. Calling C from Go means cgo, which
costs cross-compilation and build simplicity, and here it would buy nothing — the point of
this repository is to compare how each language solves the problem natively.

## Layout

```
spec/          the shared fixture and the golden output all three must produce
csharp/Settle  .NET 8 console app, with the P/Invoke path under Interop/
cpp/           C++20, headers in include/, sources in src/, plus the exported C ABI
go/            Go module, with unit tests and a golden-file test
```

## Where this comes from

I build payment systems for a living: ISO 20022 messaging, RTGS, ACH, and instant payments
under a hard ten-second deadline. Most of that code lives behind a bank's firewall, so this
repository is the part I can show.
