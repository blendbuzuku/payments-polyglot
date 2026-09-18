namespace Settle.Domain;

/// <summary>
/// Where IBAN validation comes from. The batch processor doesn't care whether the
/// answer is computed in managed code or by the native library; this is the seam
/// that lets the same run be executed either way and the outputs compared.
/// </summary>
public interface IIbanValidator
{
    bool IsValid(string iban);
}

/// <summary>The C# implementation.</summary>
public sealed class ManagedIbanValidator : IIbanValidator
{
    public bool IsValid(string iban) => Iban.IsValid(iban);
}
