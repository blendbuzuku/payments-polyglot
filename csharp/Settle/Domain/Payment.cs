namespace Settle.Domain;

/// <param name="Amount">Minor units (cents), never a floating point value.</param>
public sealed record Payment(
    string Id,
    string Debtor,
    string Creditor,
    long Amount,
    string Currency,
    bool Urgent);
