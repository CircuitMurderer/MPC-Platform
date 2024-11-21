cmake_minimum_required(VERSION 3.22)
set(CMAKE_CXX_STANDARD 11)

set(CMAKE_C_COMPILER "gcc-11")
set(CMAKE_CXX_COMPILER "g++-11")
set(CMAKE_CXX_STANDARD_REQUIRED True)

project(calculate_verifier)

set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
set(SCI_DIR "/root/install/sci/lib/cmake/SCI")

list(APPEND NEEDED_Boost_COMPONENTS
        system 
        filesystem)

find_package(Boost REQUIRED
        COMPONENTS
        ${NEEDED_Boost_COMPONENTS})

set(Boost_USE_STATIC_LIBS ON)

find_package(SCI CONFIG REQUIRED)
link_directories(${Boost_LIBRARY_DIRS})

include_directories("/root/install/sci/include")
# set_target_properties(${target} PROPERTIES LINKER_LANGUAGE CXX)

add_executable(sharer src/sharer.cpp src/utils.hpp)
target_link_libraries(
	sharer
        Boost::system
        Boost::filesystem
        SCI::SCI-FloatingPoint
        -static-libgcc
        -static-libstdc++
)

add_executable(verifier src/verifier.cpp src/utils.hpp)
target_link_libraries(
	verifier
        SCI::SCI-FloatingPoint
        -static-libgcc
        -static-libstdc++
)


