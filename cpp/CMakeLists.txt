cmake_minimum_required(VERSION 3.20)
project(settle CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

add_executable(settle
    src/main.cpp
    src/iban.cpp
    src/money.cpp
    src/processor.cpp
)

target_include_directories(settle PRIVATE include)

# The same IBAN validator, exported behind a flat C ABI so .NET can call it
# through P/Invoke: settle_iban.dll on Windows, libsettle_iban.so on Linux.
add_library(settle_iban SHARED
    src/iban.cpp
    src/iban_c.cpp
)

target_include_directories(settle_iban PUBLIC include)
target_compile_definitions(settle_iban PRIVATE SETTLE_IBAN_EXPORTS)
set_target_properties(settle_iban PROPERTIES
    CXX_VISIBILITY_PRESET hidden   # export only what the header marks
    VISIBILITY_INLINES_HIDDEN ON
)

if(MSVC)
    target_compile_options(settle PRIVATE /W4 /permissive- /EHsc)
    target_compile_options(settle_iban PRIVATE /W4 /permissive- /EHsc)
else()
    target_compile_options(settle PRIVATE -Wall -Wextra -Wpedantic)
    target_compile_options(settle_iban PRIVATE -Wall -Wextra -Wpedantic)
endif()
