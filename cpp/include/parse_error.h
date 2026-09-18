#pragma once

#include <stdexcept>

namespace settle {

// A line that cannot be read as a payment at all, as opposed to one that fails a business rule.
class ParseError : public std::runtime_error {
public:
    using std::runtime_error::runtime_error;
};

}  // namespace settle
