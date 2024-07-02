cmake_minimum_required(VERSION 3.22)
set(CMAKE_CXX_STANDARD 11)

project(verifier)

# set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -O2 -g -march=native -funroll-loops -malign-double -minline-all-stringops -flto -fuse-linker-plugin")
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
set(SCI_DIR "/root/install/sci/lib/cmake/SCI")

list(APPEND MOTION_Boost_COMPONENTS
        system 
        filesystem)

find_package(Boost REQUIRED
        COMPONENTS
        ${MOTION_Boost_COMPONENTS})

set(Boost_USE_STATIC_LIBS ON)

find_package(SCI CONFIG REQUIRED)
include_directories("/root/install/sci/include")

add_executable(verifier src/verifier.cpp src/utils.hpp)
target_link_libraries(
	verifier
        SCI::SCI-FloatingPoint
        # PRIVATE cppzmq ${CMAKE_THREAD_LIBS_INIT}
        # SCI::SCI-BuildingBlocks
        # SCI::SCI-utils
        # SCI::SCI-LinearOT
        # SCI::SCI-NonLinear
)

add_executable(sharer src/sharer.cpp src/utils.hpp)
target_link_libraries(
	sharer
        Boost::system
        Boost::filesystem
        SCI::SCI-FloatingPoint
)
