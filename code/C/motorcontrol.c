#include <stdio.h>
#ifdef PICOW
#include "pico/cyw43_arch.h"
#endif
#include "pico/multicore.h"
#include "pico/stdlib.h"
#include "motorcontrol.h"

//===================================
/*
 * Accepts values via the fifo queue from core-0; usually the next 
 * target focus-position from the * SN command. 
 * When the new value is SIGNAL_STOP, treat as Emergency Stop (from
 * the FQ command). 
 * When the new value is SIGNAL_GO, it is from the FG "Go" command.
 * Also contains optional Intervalometer functions.
 */
void motorControl() {

  stepping = false;
  imagingStarted = false;
  shutterClickRequired = false;
  shutterOpen = false;

  while (true) {

    if (multicore_fifo_rvalid){
      if (multicore_fifo_pop_timeout_us(10, &newInput)){  // microseconds! 
        handleNewInput();
      }
    }
    
    // Do we have one (but not both) buttons pressed?
    if (gpio_get(SW_INCREASE) != gpio_get(SW_DECREASE) ){
      allowManualFocus();
      //Best practise is not to mix manual- and auto- focus but that's up to the user.
    }

    if (stepping) {
      handleStepping();
    } else {
      sleep_ms(50);
      #ifdef IMAGING
      checkImagingSwitch();
      #endif
    }
  }
}

/*
* React to the manual focus buttons being pressed.
*/
void allowManualFocus(){
  
  stepping = true;
  if (gpio_get(SW_INCREASE) == sw_on){
    locationTarget = locationCurrent + 5;
  }

  if (gpio_get(SW_DECREASE) == sw_on){
    if (locationCurrent > 4){
      locationTarget = locationCurrent - 5;
    }
  }
  setStepsDirection();
	sleep_ms(10);  // otherwise it operates very quickly
}
/*
* Drive the stepper motor from locationCurrent to locationTarget.
* Gradual acceleration applied to ease the start-up and hopefully 
* avoid skipping steps, but this is not measured.
* Report back to Core-0 on current location.
*/
void handleStepping(){
  int theseSteps;
  if (locationCurrent != locationTarget) {
    
    theseSteps = stepsReqd;
    if (theseSteps > maxSteps){
      theseSteps = maxSteps;
    }
    stepsReqd -= theseSteps;
    //---------------------------------------
    //printf("at [%d] theseSteps [%d] reqd [%d]\n", locationCurrent, theseSteps, stepsReqd);
    while (theseSteps > 0) {
      theseSteps --;
      gpio_put(TMC_STEP, true);
      sleep_ms(stepDuration);
      gpio_put(TMC_STEP, false);
      sleep_ms(stepDuration);
      locationCurrent += delta;
      // gradual acceleration for stepper
      if (stepDuration > minStepDuration) {
        stepDuration -= 2;
      }
    }
    
    //---------------------------------------
    // Finished these steps, update core-0 with current location.
    if (multicore_fifo_wready){
      multicore_fifo_push_timeout_us(locationCurrent, 10);
    }

  } else {
    if (stepping){
      stepping = false;
      int shortDelay = add_alarm_in_ms(2 * 1000, disableStepper, NULL, false);
      // Now block until we are sure we update current location
      // as Main thread doesn't keep up with the stepper.
      multicore_fifo_push_blocking(locationCurrent);
    }
  }
}

// react to a new value on the FIFO queue
void handleNewInput(){
  // * * * * Emergency stop * * * *
  if (newInput == SIGNAL_STOP) { 
    locationTarget = locationCurrent;
    stepping = false;
    gpio_put(TMC_STOPPED, true);
    return;
  }

  // begin stepping
  if (newInput == SIGNAL_GO) { 
    stepping = true;
    gpio_put(TMC_STOPPED, false);
    stepDuration = maxStepDuration;
    return;
  }

  // Work-around for unknown start point & user forgot to reset 
  // focuser to zero before last power-down.
  if (newInput == SIGNAL_BACK1000) { 
    stepping = true;
    gpio_put(TMC_STOPPED, false);
    locationCurrent = 1000;
    newInput = 0;
    stepDuration = maxStepDuration;
    // don't return, continue below.
  }

  locationTarget = newInput;
  setStepsDirection();

}

void setStepsDirection(){
  if (gpio_get(TMC_STOPPED) == sw_off){
    gpio_put(TMC_STOPPED, sw_on);
  }

  if (locationTarget > locationCurrent) {
    delta = 1;
    gpio_put(TMC_DIRECTION, true);
    stepsReqd = locationTarget - locationCurrent;
  } else {
    delta = -1;
    gpio_put(TMC_DIRECTION, false);
    stepsReqd = locationCurrent - locationTarget;
  }
  //printf("Steps reqd [%d]\n", stepsReqd);
}

/*
* Pause for a few seconds before disabling the stepper.
* Prevents a "chugging" noise during manual focus as the
* stepper is no longer rapidly switched on and off. 
*/
int64_t disableStepper(alarm_id_t id, void *user_data) {
    gpio_put(TMC_STOPPED, true);
    return 0;
}

#ifdef IMAGING
//-----------------------------------------------------------------
// Optional functionality for acting as an Intervalometer
void checkImagingSwitch() {
  if (gpio_get(SW_IMAGING) == sw_on) { 
    if (!imagingStarted){
        imagingStarted = true;
        shutterClickRequired = true;
    }
    
  }
  handleImaging();
}

/*
* Only do anything if the Click required flag is set.
* Click shutter open, wait for <exposure> seconds
* Click shutter closed, wait for <pause> seconds
*/
void handleImaging() {
  if (!shutterClickRequired){
      return;  
  }

  // Always allow shutter to close, but only open if switch is on
  // as imaging can be turned off inside or outside exposure period.
  if (shutterOpen || gpio_get(SW_IMAGING) == sw_on){ 
    fireImagingShutterLED();
    shutterOpen = !shutterOpen;
    //printf("Shutter open :%b\n", shutterOpen);

    // Echo the shutter status with the on-board LED
    #if defined(PICOW)
    cyw43_arch_gpio_put(CYW43_WL_GPIO_LED_PIN, shutterOpen);
    #else
    gpio_put(LED_PIN, shutterOpen);
    #endif
  }

  // Begin an exposure and set delay for the number of seconds required.
  if (shutterOpen){
    int longDelay = add_alarm_in_ms(IMAGING_SECONDS * 1000, triggerShutter, NULL, false);
    return;
  }

 // shutter closed, is it time to stop?
  if (gpio_get(SW_IMAGING) == sw_off) { 
    imagingStarted = false;
    shutterClickRequired = false;
    return;
  }

/*
* Set next shutter Open time to a few seconds after shutter Close time, to allow camera to
* save the RAW image - which takes a short while.
*/
  if (gpio_get(SW_IMAGING) == sw_on) { 
    int shortDelay = add_alarm_in_ms(IMAGING_PAUSE * 1000, triggerShutter, NULL, false);
  }

}

// =============================
/*
* Timings adjusted from https://github.com/MATT-ER-HORN/multiCameraIrControl
* (Canon cameras only, others are available)
* original work (apparently) by Sebastien Setz which seems to be unavailable now
*/
void fireImagingShutterLED() {
  uint8_t i;
  sleep_us(200);                // seems to improve the short sleep timing
  shutterClickRequired = false;
	for (i = 0; i < 16; i++) {
		gpio_put(LED_IMAGING, true);
		sleep_us(IMAGING_LED_PULSE);
		gpio_put(LED_IMAGING, false);
    sleep_us(IMAGING_LED_PULSE);
  }

  sleep_us(IMAGING_CLICK_PAUSE);

  for (i = 0; i < 16; i++) {
		gpio_put(LED_IMAGING, true);
    sleep_us(IMAGING_LED_PULSE);
    gpio_put(LED_IMAGING, false);
    sleep_us(IMAGING_LED_PULSE);
  }
}

/* 
* Click needed to both open and close the shutter
* when its on Bulb setting for long exposures
*/
int64_t triggerShutter(alarm_id_t id, void *user_data) {
    shutterClickRequired = true;
    return 0;
}
#endif
