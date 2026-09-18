@echo off
rem ---------------------------------------------------------------------------
rem  settle - run the same payment batch processor in whichever language you pick.
rem
rem  Usage:  run.bat            menu
rem          run.bat 3          straight to Go
rem          run.bat 4 my.csv   compare all three on your own file
rem ---------------------------------------------------------------------------
setlocal enabledelayedexpansion
cd /d "%~dp0"

set "CSV=%~2"
if "%CSV%"=="" set "CSV=spec\payments.csv"
for %%F in ("%CSV%") do set "CSVFULL=%%~fF"
if not exist "%CSVFULL%" (
    echo Input file not found: %CSVFULL%
    exit /b 1
)

set "CHOICE=%~1"
if not "%CHOICE%"=="" goto dispatch

:menu
echo.
echo   settle  -  one payment batch processor, three languages
echo   input:  %CSVFULL%
echo.
echo     [1]  C#     .NET 8
echo     [2]  C++    C++20, MSVC
echo     [3]  Go
echo     [4]  All three, compared against spec\expected.txt
echo     [5]  C# calling the C++ library through P/Invoke
echo     [Q]  Quit
echo.
set "CHOICE="
set /p "CHOICE=Which one? "
if /i "%CHOICE%"=="Q" exit /b 0
if "%CHOICE%"=="" goto menu

:dispatch
echo.
if "%CHOICE%"=="1" ( call :run_csharp "%CSVFULL%" & goto done )
if "%CHOICE%"=="2" ( call :run_cpp     "%CSVFULL%" & goto done )
if "%CHOICE%"=="3" ( call :run_go      "%CSVFULL%" & goto done )
if "%CHOICE%"=="4" ( call :run_all     "%CSVFULL%" & goto done )
if "%CHOICE%"=="5" ( call :run_interop "%CSVFULL%" & goto done )
echo "%CHOICE%" is not one of the options.
if "%~1"=="" goto menu
exit /b 1

:done
exit /b %ERRORLEVEL%


rem --------------------------------------------------------------------- C# --
:run_csharp
where dotnet >nul 2>&1 || (
    echo The .NET SDK was not found. Install it from https://dotnet.microsoft.com/download
    exit /b 1
)
echo --- C# ------------------------------------------------------------------
dotnet run --project "csharp\Settle" -- "%~1"
exit /b %ERRORLEVEL%


rem -------------------------------------------------------------------- C++ --
:run_cpp
echo --- C++ -----------------------------------------------------------------
call "cpp\build.bat" >nul || (
    echo The C++ build failed. Run cpp\build.bat on its own to see why.
    exit /b 1
)
"cpp\build\settle.exe" "%~1"
exit /b %ERRORLEVEL%


rem --------------------------------------------------------------------- Go --
:run_go
call :find_go || exit /b 1
echo --- Go ------------------------------------------------------------------
"%GOEXE%" -C "go" run . "%~1"
exit /b %ERRORLEVEL%


:find_go
set "GOEXE=go"
where go >nul 2>&1 && exit /b 0
set "GOEXE=%ProgramFiles%\Go\bin\go.exe"
if exist "%GOEXE%" exit /b 0
echo Go was not found. Install it from https://go.dev/dl/ , or open a new terminal
echo if you installed it after this window was opened.
exit /b 1


rem ----------------------------------------------------------------- interop --
:run_interop
where dotnet >nul 2>&1 || ( echo The .NET SDK was not found. & exit /b 1 )
echo --- C# with the native C++ validator ------------------------------------
call "cpp\build.bat" >nul || (
    echo Could not build the native library. Run cpp\build.bat to see why.
    exit /b 1
)
dotnet build "csharp\Settle" -v q --nologo >nul || exit /b 1
copy /y "cpp\build\settle_iban.dll" "csharp\Settle\bin\Debug\net8.0\" >nul
"csharp\Settle\bin\Debug\net8.0\settle.exe" "%~1" --native
exit /b %ERRORLEVEL%


rem ---------------------------------------------------------------- all three --
:run_all
set "OUT=%TEMP%\settle-run"
if not exist "%OUT%" mkdir "%OUT%"
set "FAILED="

echo Running all three against %~1 ...
echo.

call :run_csharp "%~1" > "%OUT%\csharp.txt" 2>&1
call :compare "C#" "%OUT%\csharp.txt"

call :run_cpp "%~1" > "%OUT%\cpp.txt" 2>&1
call :compare "C++" "%OUT%\cpp.txt"

call :run_go "%~1" > "%OUT%\go.txt" 2>&1
call :compare "Go" "%OUT%\go.txt"

echo.
if defined FAILED (
    echo Not every language matched. The reports are in %OUT%
    exit /b 1
)
echo All three produced the same report.
echo Reports: %OUT%
exit /b 0


:compare
rem %1 = label, %2 = output file. Strips the "--- label ---" banner line, then
rem compares line by line with PowerShell so CRLF and LF endings both work:
rem Go writes LF, while C# and C++ write CRLF on Windows.
powershell -NoProfile -Command ^
  "$want = Get-Content 'spec\expected.txt';" ^
  "$got = Get-Content '%~2' | Where-Object { $_ -notmatch '^--- ' };" ^
  "if (Compare-Object -CaseSensitive $want $got) { exit 1 } else { exit 0 }"
if errorlevel 1 (
    echo   %~1 : DIFFERS from spec\expected.txt  ^(see %~2^)
    set "FAILED=1"
) else (
    echo   %~1 : matches spec\expected.txt
)
exit /b 0
