#pragma once

#include <iosfwd>
#include <map>
#include <memory>
#include <string>
#include <unordered_set>
#include <vector>

#include "money.h"
#include "rail.h"

namespace settle {

struct Payment {
    std::string id;
    std::string debtor;
    std::string creditor;
    Cents amount = 0;
    std::string currency;
    bool urgent = false;
};

struct RailTotals {
    int count = 0;
    Cents volume = 0;
    Cents fees = 0;
};

// Reads a clearing file, applies the scheme rules, routes each accepted payment
// and prints the report. Rejection reasons are checked in a fixed order and the
// first failure wins, so a line always has exactly one reason.
class BatchProcessor {
public:
    BatchProcessor();

    void run(std::istream& input, std::ostream& output);

private:
    void process(const std::string& line, std::ostream& output);
    Payment readPayment(const std::vector<std::string>& fields) const;
    std::string_view check(const Payment& payment);
    const Rail& route(const Payment& payment) const;
    void reject(std::string_view id, std::string_view reason, std::ostream& output);
    void post(const Payment& payment, Cents fee);

    std::vector<std::unique_ptr<Rail>> rails_;         // owning, freed when the processor dies
    std::map<std::string_view, RailTotals> totals_;
    std::unordered_set<std::string> seenIds_;
    std::map<std::string, Cents> positions_;           // ordered, so the report needs no sort
    int accepted_ = 0;
    int rejected_ = 0;
};

}  // namespace settle
