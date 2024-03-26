# Moonlyte
An electronic focuser for Skywatcher Skymax 180 telescopes, emulating the Moonlite protocol

This project is licensed under the terms of the MIT license.

Written in GO / Tinygo as I had not tried the language much. N.B. you will have to adjust settings.json
as it refers to a location for the GOROOT which is on my disk, not yours.

Uses a Pico microcontroller instead of the traditional Arduino, 
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

The physical parts could be improved too. The body is quite fiddly to assemble and prevents manual focusing, but as a first pass its working well.

2023-12-15 Version 2 hardware & slight update for software (mainly prevent "chugging" where the manual input causes the Stepper to be enabled and disabled rapidly)
V2 came about to make room for an Off-Axis Guider.

2024-01-24 Added Intervalometer and bigger box to hold the manual focus and Intervalometer on-off switches

2024-03-26 Added C version of the focuser code for proper multi-threading, reorganised code directories


<pre>
3d printing files Version 1 hardware:
FocuserV5.scad        Design file for Focuser body (OpenSCAD)
FocuserBox.scad       Design file for Box to hold electronics (OpenSCAD)
F.* STL files for the focuser body
</pre>
<pre>
3d printing files Version 2 hardware:
FocuserV6.scad        Design file for Focuser body (OpenSCAD)
FocuserBox.scad       (same as before)
F.* STL files for the focuser body
</pre>

<pre>
Driver module from https://www.amazon.co.uk/dp/B0893DPSJL 
Stepper motor from https://www.amazon.co.uk/dp/B08DCGHLYP. 
</pre>
