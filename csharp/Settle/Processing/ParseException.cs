namespace Settle.Domain;

/// <summary>A line that cannot be read as a payment at all, as opposed to one that fails a business rule.</summary>
public sealed class ParseException(string message) : Exception(message);
