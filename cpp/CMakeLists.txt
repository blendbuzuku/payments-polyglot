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

if(MSVC)
    target_compile_options(settle PRIVATE /W4 /permissive- /EHsc)
else()
    target_compile_options(settle PRIVATE -Wall -Wextra -Wpedantic)
endif()
