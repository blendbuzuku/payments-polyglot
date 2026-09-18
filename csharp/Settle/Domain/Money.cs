using System.Globalization;

namespace Settle.Domain;

/// <summary>
/// Money is held as integer minor units (cents). Floating point is never used:
/// a binary double cannot represent 0.10 exactly, and a clearing file must balance
/// to the cent. Core banking systems store minor units for the same reason.
/// </summary>
public static class Money
{
    /// <summary>Parses "1250.00", "300.5" or "-50" into cents. Throws <see cref="ParseException"/> otherwise.</summary>
    public static long Parse(string text)
    {
        int i = 0;
        bool negative = text.StartsWith('-');
        if (negative) i++;

        int wholeStart = i;
        long whole = 0;
        while (i < text.Length && char.IsAsciiDigit(text[i]))
        {
            whole = whole * 10 + (text[i] - '0');
            i++;
        }
        if (i == wholeStart) throw new ParseException($"amount has no digits: '{text}'");

        long fraction = 0;
        if (i < text.Length && text[i] == '.')
        {
            int fractionStart = ++i;
            while (i < text.Length && char.IsAsciiDigit(text[i]))
            {
                fraction = fraction * 10 + (text[i] - '0');
                i++;
            }
            int digits = i - fractionStart;
            if (digits is 0 or > 2) throw new ParseException($"amount needs 1 or 2 decimals: '{text}'");
            if (digits == 1) fraction *= 10;
        }

        if (i != text.Length) throw new ParseException($"unexpected characters in amount: '{text}'");

        long cents = whole * 100 + fraction;
        return negative ? -cents : cents;
    }

    /// <summary>Formats cents as "1250.00", with a leading minus only when negative.</summary>
    public static string Format(long cents)
    {
        long absolute = Math.Abs(cents);
        return string.Create(
            CultureInfo.InvariantCulture,
            $"{(cents < 0 ? "-" : "")}{absolute / 100}.{absolute % 100:D2}");
    }
}
