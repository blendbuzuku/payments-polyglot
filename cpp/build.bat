@echo off
rem Builds settle.exe and the settle_iban shared library with MSVC, from any shell.
rem It locates Visual Studio itself, so you do not need the Developer prompt.
setlocal

set "VSWHERE=%ProgramFiles(x86)%\Microsoft Visual Studio\Installer\vswhere.exe"
if not exist "%VSWHERE%" (
    echo Could not find vswhere.exe. Is Visual Studio installed?
    exit /b 1
)

for /f "usebackq tokens=*" %%i in (`"%VSWHERE%" -latest -products * -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 -property installationPath`) do set "VSPATH=%%i"
if not defined VSPATH (
    echo Visual Studio was found, but without the "Desktop development with C++" workload.
    echo Add it in the Visual Studio Installer, then run this again.
    exit /b 1
)

rem stderr is silenced too: vcvars64 prints a harmless "vswhere.exe is not recognized"
rem when it is called from outside the Developer prompt. Failures are caught below.
call "%VSPATH%\VC\Auxiliary\Build\vcvars64.bat" >nul 2>&1
if errorlevel 1 (
    echo Could not initialise the MSVC environment.
    exit /b 1
)

cd /d "%~dp0"
if not exist build mkdir build

echo Building settle.exe ...
cl /nologo /std:c++20 /EHsc /W4 /permissive- /Iinclude ^
   src\main.cpp src\iban.cpp src\money.cpp src\processor.cpp ^
   /Fe:build\settle.exe /Fo:build\ || exit /b 1

echo Building settle_iban.dll ...
cl /nologo /LD /std:c++20 /EHsc /W4 /permissive- /DSETTLE_IBAN_EXPORTS /Iinclude ^
   src\iban.cpp src\iban_c.cpp ^
   /Fe:build\settle_iban.dll /Fo:build\ || exit /b 1

echo.
echo Built cpp\build\settle.exe and cpp\build\settle_iban.dll
echo Run: cpp\build\settle.exe spec\payments.csv
