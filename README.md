# Moonlyte
An electronic focuser for Skywatcher Skymax 180 telescopes, emulating the Moonlite protocol

Written in GO / Tinygo as I had not tried the language much.  Uses a Pico microcontroller instead of the traditional Arduino, 
again because I had not used it and it was just collecting dust.  

This implements a basic subset of the Moonlite protocol which I found documented at 
https://indilib.org/media/kunena/attachments/1/HighResSteppermotor107.pdf
(Plus a "debug" output switch.  This is non-standard and should only be used to resolve problems as
it will send text to the host which the host is not expecting.)

Sadly, there is no non-volatile memory on the Pico, so the device will not remember where it was left if you power-off.
I plan to run the focuser back to position zero before power-off as my Host software (Ain Imager & Indigo Framework 
https://www.indigo-astronomy.org/index.html) does not let me use negative numbers for focuser position.  Maybe that's true of all focusers, don't know.

Even more sadly, Tinygo does not implement true concurrency - I had hoped to use the dual processor cores of the Pico, 
with one routine for Comms to the host and the other to drive the stepper but have had to interleave stepping with comms.  
A suitable future update would be to try using the PIO to drive the stepper instead of a separate core.
