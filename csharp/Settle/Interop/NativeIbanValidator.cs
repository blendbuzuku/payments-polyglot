using System.Runtime.InteropServices;
using Settle.Domain;

namespace Settle.Interop;

/// <summary>
/// IBAN validation performed by the C++ library through P/Invoke.
///
/// The runtime maps the library name onto the platform's convention: settle_iban.dll
/// on Windows, libsettle_iban.so on Linux, libsettle_iban.dylib on macOS. It must sit
/// next to the executable or on the loader's search path, and it must match the
/// process architecture: a 64-bit process cannot load a 32-bit library.
/// </summary>
public sealed class NativeIbanValidator : IIbanValidator
{
    public bool IsValid(string iban) => NativeMethods.IbanIsValid(iban) == 1;

    /// <summary>
    /// Forces the library to load now, so a missing or mismatched build fails here
    /// with a clear message instead of midway through a batch.
    /// </summary>
    public static NativeIbanValidator Load()
    {
        var validator = new NativeIbanValidator();
        _ = validator.IsValid("XK051212012345678906");
        return validator;
    }

    private static class NativeMethods
    {
        private const string Library = "settle_iban";

        // Cdecl matches the C compiler's default on every platform we build for.
        // LPUTF8Str hands the native side a null-terminated UTF-8 buffer; the default
        // marshalling for a C# string would otherwise be UTF-16, which char* is not.
        [DllImport(Library, EntryPoint = "settle_iban_is_valid", CallingConvention = CallingConvention.Cdecl)]
        internal static extern int IbanIsValid([MarshalAs(UnmanagedType.LPUTF8Str)] string iban);
    }
}
