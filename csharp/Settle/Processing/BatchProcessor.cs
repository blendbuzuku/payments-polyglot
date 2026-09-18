using Settle.Domain;

namespace Settle.Processing;

/// <summary>
/// Reads a clearing file, applies the scheme rules, routes each accepted payment
/// and prints the report. Rejection reasons are checked in a fixed order and the
/// first failure wins, so a line always has exactly one reason.
/// </summary>
public sealed class BatchProcessor
{
    private const string SupportedCurrency = "EUR";
    private const long RtgsThreshold = 1_000_000; // 10,000.00

    private readonly Rail[] _rails = [new BookRail(), new AchRail(), new RtgsRail()];
    private readonly Dictionary<string, RailTotals> _totals;
    private readonly HashSet<string> _seenIds = [];
    private readonly Ledger _ledger = new();
    private readonly IIbanValidator _ibans;

    private int _accepted;
    private int _rejected;

    /// <param name="ibans">
    /// Managed validation by default; pass <see cref="Interop.NativeIbanValidator"/>
    /// to have the C++ library answer instead. The report must be identical either way.
    /// </param>
    public BatchProcessor(IIbanValidator? ibans = null)
    {
        _ibans = ibans ?? new ManagedIbanValidator();
        _totals = _rails.ToDictionary(rail => rail.Name, _ => new RailTotals());
    }

    public void Run(TextReader input, TextWriter output)
    {
        output.WriteLine("== Payments ==");

        bool header = true;
        for (string? line = input.ReadLine(); line is not null; line = input.ReadLine())
        {
            if (header) { header = false; continue; }
            if (line.Length == 0) continue;

            Process(line.TrimEnd('\r'), output);
        }

        WriteRails(output);
        WritePositions(output);
        WriteCheck(output);
    }

    private void Process(string line, TextWriter output)
    {
        string[] fields = line.Split(',');

        Payment payment;
        try
        {
            payment = ReadPayment(fields);
        }
        catch (ParseException)
        {
            Reject(fields.Length > 0 ? fields[0] : "", "MALFORMED", output);
            return;
        }

        string? reason = Check(payment);
        if (reason is not null)
        {
            Reject(payment.Id, reason, output);
            return;
        }

        Rail rail = Route(payment);
        long fee = rail.Fee(payment.Amount);
        _ledger.Post(payment, fee);

        RailTotals totals = _totals[rail.Name];
        totals.Count++;
        totals.Volume += payment.Amount;
        totals.Fees += fee;

        _accepted++;
        output.WriteLine($"{payment.Id} ACCEPTED {rail.Name} {Money.Format(payment.Amount)} fee={Money.Format(fee)}");
    }

    private static Payment ReadPayment(string[] fields)
    {
        if (fields.Length != 6) throw new ParseException($"expected 6 fields, got {fields.Length}");

        return new Payment(
            Id: fields[0],
            Debtor: fields[1],
            Creditor: fields[2],
            Amount: Money.Parse(fields[3]),
            Currency: fields[4],
            Urgent: ParseFlag(fields[5]));
    }

    private static bool ParseFlag(string text) => text switch
    {
        "true" => true,
        "false" => false,
        _ => throw new ParseException($"urgent must be true or false: '{text}'")
    };

    /// <summary>The first rule that fails, or null when the payment is good.</summary>
    private string? Check(Payment payment)
    {
        if (!_seenIds.Add(payment.Id)) return "DUPLICATE_ID";
        if (!_ibans.IsValid(payment.Debtor) || !_ibans.IsValid(payment.Creditor)) return "INVALID_IBAN";
        if (payment.Debtor == payment.Creditor) return "SAME_ACCOUNT";
        if (payment.Amount <= 0) return "INVALID_AMOUNT";
        if (payment.Currency != SupportedCurrency) return "UNSUPPORTED_CURRENCY";
        return null;
    }

    private Rail Route(Payment payment)
    {
        // Same bank: the money never leaves our books, whatever the sender asked for.
        if (Iban.BankCode(payment.Debtor) == Iban.BankCode(payment.Creditor)) return _rails[0];
        if (payment.Urgent || payment.Amount >= RtgsThreshold) return _rails[2];
        return _rails[1];
    }

    private void Reject(string id, string reason, TextWriter output)
    {
        output.WriteLine($"{id} REJECTED {reason}");
        _rejected++;
    }

    private void WriteRails(TextWriter output)
    {
        output.WriteLine("== Rails ==");
        foreach (Rail rail in _rails)
        {
            RailTotals totals = _totals[rail.Name];
            output.WriteLine(
                $"{rail.Name} count={totals.Count} volume={Money.Format(totals.Volume)} fees={Money.Format(totals.Fees)}");
        }
    }

    private void WritePositions(TextWriter output)
    {
        output.WriteLine("== Positions ==");
        foreach ((string account, long cents) in _ledger.Positions)
            output.WriteLine($"{account} {Money.Format(cents)}");
    }

    private void WriteCheck(TextWriter output)
    {
        long sum = _ledger.Sum;
        output.WriteLine("== Check ==");
        output.WriteLine(
            $"accepted={_accepted} rejected={_rejected} sum={Money.Format(sum)} {(sum == 0 ? "OK" : "MISMATCH")}");
    }

    private sealed class RailTotals
    {
        public int Count;
        public long Volume;
        public long Fees;
    }
}
