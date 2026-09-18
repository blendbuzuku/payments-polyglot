using Settle.Domain;
using Settle.Interop;
using Settle.Processing;

string[] files = args.Where(a => !a.StartsWith("--")).ToArray();
bool useNative = args.Contains("--native");

if (files.Length < 1)
{
    Console.Error.WriteLine("usage: settle <payments.csv> [--native]");
    Console.Error.WriteLine("  --native   validate IBANs through the C++ library instead of managed code");
    return 1;
}

if (!File.Exists(files[0]))
{
    Console.Error.WriteLine($"cannot open {files[0]}");
    return 1;
}

IIbanValidator validator;
if (useNative)
{
    try
    {
        validator = NativeIbanValidator.Load();
    }
    catch (DllNotFoundException)
    {
        Console.Error.WriteLine(
            "cannot load the native library. Build it first (see cpp/README or the interop section of the main README), " +
            "then put settle_iban.dll / libsettle_iban.so next to this executable.");
        return 1;
    }
    catch (EntryPointNotFoundException)
    {
        Console.Error.WriteLine(
            "the library loaded but settle_iban_is_valid was not found in it. " +
            "That usually means it was built without extern \"C\", so the symbol is mangled.");
        return 1;
    }
}
else
{
    validator = new ManagedIbanValidator();
}

using StreamReader input = File.OpenText(files[0]);
new BatchProcessor(validator).Run(input, Console.Out);
return 0;
