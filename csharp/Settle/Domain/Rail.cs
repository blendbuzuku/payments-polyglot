namespace Settle.Domain;

/// <summary>A clearing rail: what a payment travels on, and what that costs.</summary>
public abstract class Rail
{
    public abstract string Name { get; }

    /// <summary>Fee in cents for an amount in cents.</summary>
    public abstract long Fee(long amount);
}

/// <summary>Both parties bank with us, so the transfer never leaves the books. Free.</summary>
public sealed class BookRail : Rail
{
    public override string Name => "BOOK";
    public override long Fee(long amount) => 0;
}

/// <summary>Batch clearing: cheap, settled in the next cycle.</summary>
public sealed class AchRail : Rail
{
    public override string Name => "ACH";
    public override long Fee(long amount) => 20;
}

/// <summary>Real-time gross settlement: 1.50 plus 0.01% of the amount, rounded half-up.</summary>
public sealed class RtgsRail : Rail
{
    public override string Name => "RTGS";

    // 0.01% of `amount` is amount/10000 cents; adding 5000 before the integer
    // division rounds half-up without ever touching a double.
    public override long Fee(long amount) => 150 + (amount + 5000) / 10000;
}
