using Settle.Processing;

if (args.Length < 1)
{
    Console.Error.WriteLine("usage: settle <payments.csv>");
    return 1;
}

if (!File.Exists(args[0]))
{
    Console.Error.WriteLine($"cannot open {args[0]}");
    return 1;
}

using StreamReader input = File.OpenText(args[0]);
new BatchProcessor().Run(input, Console.Out);
return 0;
