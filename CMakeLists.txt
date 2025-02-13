cmake_minimum_required(VERSION 3.10)

project(berenc C)

set(CMAKE_C_STANDARD 99)
set(CMAKE_C_STANDARD_REQUIRED ON)

set(SOURCES main.c ber.c)

# Ensure MSVC is used
if(MSVC)
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} /W4 /Zi")  # Enable debugging and warnings
    set(CMAKE_EXE_LINKER_FLAGS "${CMAKE_EXE_LINKER_FLAGS} /DEBUG")
endif()

# Detect MSVC and set manually
if(WIN32)
    set(CMAKE_C_COMPILER "cl")
endif()

add_executable(berenc ${SOURCES})

set(CMAKE_BUILD_TYPE Debug)
