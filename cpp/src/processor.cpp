#include "processor.h"

#include <istream>
#include <ostream>
#include <sstream>
#include <utility>

#include "iban.h"
#include "parse_error.h"

namespace settle {
namespace {

constexpr std::string_view kSupportedCurrency = "EUR";
constexpr Cents kRtgsThreshold = 1'000'000;  // 10,000.00

std::vector<std::string> split(const std::string& line, char separator) {
    std::vector<std::string> fields;
    std::istringstream stream(line);
    std::string field;
    while (std::getline(stream, field, separator)) fields.push_back(field);
    if (!line.empty() && line.back() == separator) fields.emplace_back();
    return fields;
}

bool parseFlag(const std::string& text) {
    if (text == "true") return true;
    if (text == "false") return false;
    throw ParseError("urgent must be true or false: '" + text + "'");
}

}  // namespace

BatchProcessor::BatchProcessor() {
    rails_.push_back(std::make_unique<BookRail>());
    rails_.push_back(std::make_unique<AchRail>());
    rails_.push_back(std::make_unique<RtgsRail>());
    for (const auto& rail : rails_) totals_[rail->name()] = RailTotals{};
}

void BatchProcessor::run(std::istream& input, std::ostream& output) {
    output << "== Payments ==\n";

    std::string line;
    bool header = true;
    while (std::getline(input, line)) {
        if (!line.empty() && line.back() == '\r') line.pop_back();  // a file written on Windows
        if (header) {
            header = false;
            continue;
        }
        if (line.empty()) continue;
        process(line, output);
    }

    output << "== Rails ==\n";
    for (const auto& rail : rails_) {
        const RailTotals& totals = totals_[rail->name()];
        output << rail->name() << " count=" << totals.count
               << " volume=" << formatAmount(totals.volume)
               << " fees=" << formatAmount(totals.fees) << '\n';
    }

    output << "== Positions ==\n";
    Cents sum = 0;
    for (const auto& [account, cents] : positions_) {
        output << account << ' ' << formatAmount(cents) << '\n';
        sum += cents;
    }

    output << "== Check ==\n"
           << "accepted=" << accepted_ << " rejected=" << rejected_
           << " sum=" << formatAmount(sum) << (sum == 0 ? " OK" : " MISMATCH") << '\n';
}

void BatchProcessor::process(const std::string& line, std::ostream& output) {
    const std::vector<std::string> fields = split(line, ',');

    Payment payment;
    try {
        payment = readPayment(fields);
    } catch (const ParseError&) {
        reject(fields.empty() ? std::string_view{} : std::string_view{fields[0]}, "MALFORMED", output);
        return;
    }

    if (const std::string_view reason = check(payment); !reason.empty()) {
        reject(payment.id, reason, output);
        return;
    }

    const Rail& rail = route(payment);
    const Cents fee = rail.fee(payment.amount);
    post(payment, fee);

    RailTotals& totals = totals_[rail.name()];
    ++totals.count;
    totals.volume += payment.amount;
    totals.fees += fee;

    ++accepted_;
    output << payment.id << " ACCEPTED " << rail.name() << ' ' << formatAmount(payment.amount)
           << " fee=" << formatAmount(fee) << '\n';
}

Payment BatchProcessor::readPayment(const std::vector<std::string>& fields) const {
    if (fields.size() != 6) {
        throw ParseError("expected 6 fields, got " + std::to_string(fields.size()));
    }

    Payment payment;
    payment.id = fields[0];
    payment.debtor = fields[1];
    payment.creditor = fields[2];
    payment.amount = parseAmount(fields[3]);
    payment.currency = fields[4];
    payment.urgent = parseFlag(fields[5]);
    return payment;
}

// The first rule that fails, or an empty view when the payment is good.
std::string_view BatchProcessor::check(const Payment& payment) {
    if (!seenIds_.insert(payment.id).second) return "DUPLICATE_ID";
    if (!isValidIban(payment.debtor) || !isValidIban(payment.creditor)) return "INVALID_IBAN";
    if (payment.debtor == payment.creditor) return "SAME_ACCOUNT";
    if (payment.amount <= 0) return "INVALID_AMOUNT";
    if (payment.currency != kSupportedCurrency) return "UNSUPPORTED_CURRENCY";
    return {};
}

const Rail& BatchProcessor::route(const Payment& payment) const {
    // Same bank: the money never leaves our books, whatever the sender asked for.
    if (bankCode(payment.debtor) == bankCode(payment.creditor)) return *rails_[0];
    if (payment.urgent || payment.amount >= kRtgsThreshold) return *rails_[2];
    return *rails_[1];
}

void BatchProcessor::reject(std::string_view id, std::string_view reason, std::ostream& output) {
    output << id << " REJECTED " << reason << '\n';
    ++rejected_;
}

// Double entry: the sender pays amount plus fee, the beneficiary receives the amount,
// and FEES takes the difference, so the batch nets to exactly zero.
void BatchProcessor::post(const Payment& payment, Cents fee) {
    positions_[payment.debtor] -= payment.amount + fee;
    positions_[payment.creditor] += payment.amount;
    positions_["FEES"] += fee;
}

}  // namespace settle
