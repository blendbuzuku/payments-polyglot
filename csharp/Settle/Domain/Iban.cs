namespace Settle.Domain;

/// <summary>IBAN validation per ISO 13616: structure plus the ISO 7064 MOD-97-10 check.</summary>
public static class Iban
{
    public static bool IsValid(string iban)
    {
        if (iban.Length is < 15 or > 34) return false;

        for (int i = 0; i < iban.Length; i++)
        {
            char c = iban[i];
            bool ok = i switch
            {
                < 2 => char.IsAsciiLetterUpper(c),               // country code
                < 4 => char.IsAsciiDigit(c),                     // check digits
                _ => char.IsAsciiLetterUpper(c) || char.IsAsciiDigit(c)
            };
            if (!ok) return false;
        }

        // Move the first four characters to the end, map A=10..Z=35, then take mod 97.
        // The rearranged number is far too large for any integer type, so the remainder
        // is folded in digit by digit.
        int remainder = 0;
        for (int k = 0; k < iban.Length; k++)
        {
            char c = iban[(k + 4) % iban.Length];
            remainder = char.IsAsciiDigit(c)
                ? (remainder * 10 + (c - '0')) % 97
                : (remainder * 100 + (c - 'A' + 10)) % 97;
        }

        return remainder == 1;
    }

    /// <summary>The four-digit bank code: characters 5-8 of the IBAN.</summary>
    public static string BankCode(string iban) => iban.Substring(4, 4);
}
