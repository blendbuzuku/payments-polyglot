using Settle.Domain;

namespace Settle.Processing;

/// <summary>
/// Double-entry positions for one batch. Every accepted payment debits the sender for
/// amount plus fee, credits the beneficiary with the amount, and credits FEES with the fee,
/// so the batch nets to exactly zero. That total is the end-of-day check.
/// </summary>
public sealed class Ledger
{
    public const string FeeAccount = "FEES";

    // Ordinal comparison so the sort order matches std::map in C++ and sort.Strings in Go.
    private readonly SortedDictionary<string, long> _positions = new(StringComparer.Ordinal);

    public void Post(Payment payment, long fee)
    {
        Add(payment.Debtor, -(payment.Amount + fee));
        Add(payment.Creditor, payment.Amount);
        Add(FeeAccount, fee);
    }

    private void Add(string account, long cents) =>
        _positions[account] = _positions.GetValueOrDefault(account) + cents;

    public IEnumerable<KeyValuePair<string, long>> Positions => _positions;

    /// <summary>Zero for a balanced batch. Anything else means a leg went missing.</summary>
    public long Sum => _positions.Values.Sum();
}
