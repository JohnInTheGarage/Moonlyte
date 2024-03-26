cmake_minimum_required(VERSION 3.13)

include(pico_sdk_import.cmake)

project(pico_moonlyte C CXX ASM)
set(CMAKE_C STANDARD 11)
set(CMAKE_CXX_STANDARD 17)

pico_sdk_init()

add_executable(focuser
    main.c 
    motorcontrol.c
    commandhandler.c
)

pico_enable_stdio_usb(focuser 1)
pico_enable_stdio_uart(focuser 0)
pico_add_extra_outputs(focuser)

target_include_directories(focuser PRIVATE ${CMAKE_CURRENT_LIST_DIR} )

if(PICO_BOARD STREQUAL "pico_w")
    target_link_libraries(
        focuser 
        pico_cyw43_arch_lwip_threadsafe_background 
        pico_stdlib 
        pico_multicore 
        hardware_adc
    )
    add_definitions(-DPICOW)
else ()    
    target_link_libraries(
        focuser 
        pico_stdlib 
        pico_multicore 
        hardware_adc
    )
endif()

if(IMAGING STREQUAL "y")
    add_definitions(-DIMAGING)
endif()
