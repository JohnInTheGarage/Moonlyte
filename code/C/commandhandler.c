#include <stdio.h>
#include <string.h>
#include "pico/stdlib.h"
#include "pico/multicore.h"
#include "commandhandler.h"

//===================================
void actionCommand(struct FocuserCommand c){

  if (strcmp(c.action, "FQ") == 0){
    emergencyStop();
  }
  if (strcmp(c.action, "FG") == 0){
    goToTarget();
  }
  if (strcmp(c.action, "GH") == 0){
    sendHalfstepStatus();
  }
  if (strcmp(c.action, "GC") == 0){
    sendTempCoefficient();
  }
  if (strcmp(c.action, "GI") == 0){
    sendMovingStatus();
  }
  if (strcmp(c.action, "GN") == 0){
    sendTargetPosition();
  }
  if (strcmp(c.action, "GP") == 0){
    sendCurrentPosition();
  }
  if (strcmp(c.action, "GT") == 0){
    sendTemperature();
  }
  if (strcmp(c.action, "GV") == 0){
    sendVersion();
  }
  if (strcmp(c.action, "GD") == 0){
    sendStepDelay();
  }
  if (strcmp(c.action, "SD") == 0){
    setStepDelay(c.value);
  }
  // >>>>>>>>>>>>>>>> non-standard!  May be removed
  if (strcmp(c.action, "SF") == 0){
    hijackFullStepMode();
  }
  // >>>>>>>>>>>>>>>> non-standard!  May be removed
  if (strcmp(c.action, "SH") == 0){
    hijackHalfStepMode();
  }
  if (strcmp(c.action, "SN") == 0){
    setNewTarget(c.value);
  }
  if (strcmp(c.action, "YB") == 0){
    setBacklash();
  }
  if (strcmp(c.action, "ZB") == 0){
    sendBacklash();
  }
  if (strcmp(c.action, "ZT") == 0){
    sendMaxSteps();
  }
  if (strcmp(c.action, "ZA") == 0){
    sendAverageTemp();
  }
}
// ======================
void sendAverageTemp() {

}

// ======================
void sendMaxSteps() {

}

// ======================
void sendBacklash() {

}

// ======================
void setBacklash() {

}

// ======================
void setNewTarget(int value){
  locationRequired = value;
  passToCore1(value);
}

// ======================
/*
 * Since the TMC2209 is already micro-stepping,
 * Full step and Half Step commands are re-purposed
 * to overcome lack of known start-point (only needed
 * if you forget to return the focuser to zero when closing down).
 */
void hijackFullStepMode(){ 
  // Restore normal direction
  // Nothing needed here, but it allows the host to know
  // that we are "back on full step mode"
  // in case we need to backup another 1000 steps.
  
}

// ======================
/*
 * Reset the "zero" position.  Go backwards 1000 steps and
 * change location to zero
 * The User is expected to be monitoring progress carefully!
 */
void hijackHalfStepMode(){
  passToCore1(SIGNAL_BACK1000);
}

// =======================
void sendStepDelay() { 
  respondAsHex2(unusedStepDelay); 
}

// =======================
void setStepDelay(int value) { 
  unusedStepDelay = (uint16_t)value; 
}

// ======================
void sendVersion(){
  respond("10");
}

// =========================
void sendTempCoefficient() { 
  respond("00"); 
}

// ======================
void sendTemperature(){
  respondAsHex4(getTemperature());
}

// ======================
void sendCurrentPosition(){
  respondAsHex4(locationReported);
}

// ======================
void sendTargetPosition(){
  respondAsHex4(locationRequired);
}

// ======================
void sendMovingStatus(){
  if (locationRequired != locationReported){
    respond("01");
  } else {
    respond("00");
  }
}

// ======================
void sendHalfstepStatus(){
  respond("00");
}

// ======================
void goToTarget(){
  passToCore1(SIGNAL_GO);
}

// ======================
void emergencyStop(){
  passToCore1(SIGNAL_STOP);
}

void respond(char *text){
  printf(text);
  printf("#\0");
}

void respondAsHex2(int num){
  printf("%02X#\0", num);
}

void respondAsHex4(int num){
  printf("%04X#\0", num);
}

void passToCore1(int item){
  while (!multicore_fifo_wready){
    sleep_ms(10);
  }
  multicore_fifo_push_blocking(item);
}
