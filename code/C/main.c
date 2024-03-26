#include <string.h>
#include <stdlib.h>
#include <stdio.h>
#include "pico/multicore.h"
#include "pico/stdlib.h"
#include "hardware/adc.h"
#include "commandhandler.h"
#ifdef PICOW
#include "pico/cyw43_arch.h"
#endif
/*==============================================
* C version of code enables proper multi-threading as both RP2040 cores
* are available for use (unlike in tinygo).  Core-0 handles comms with Host PC
* while core-1 drives the focuser stepper-motor and does the optional
* Intervalometer functions.
*
* Multi-core based on ideas from https://github.com/bnielsen1965 via
* https://www.youtube.com/watch?v=Nz68nY5hzxw
*
* for PICO-W use : cmake -DPICO_BOARD=pico_w ..
* for PICO use   : cmake  ..
* for Intervalometer functions include -DIMAGING=y
==============================================*/

// GPIO Pins & signal values for Stepper control
const uint8_t TMC_STEP = 16;
const uint8_t TMC_DIRECTION = 17;
// Actually its the ENABLE pin but STOPPED reads better.
const uint8_t TMC_STOPPED = 18;
const uint8_t LED_PIN = 25; // for PICO not PICO-W

const uint32_t SIGNAL_STOP = 0x0F0000;
const uint32_t SIGNAL_GO = 0x0A0000;
const uint32_t SIGNAL_BACK1000 = 0x0B0000;

// For manual focus buttons
const uint8_t SW_DECREASE = 19;
const uint8_t SW_INCREASE = 20;

// For Intervalometer
#ifdef IMAGING
const uint8_t SW_IMAGING = 10;
const uint8_t LED_IMAGING = 11;
const uint8_t IMAGING_SECONDS = 60;
const uint8_t IMAGING_PAUSE = 8;
const uint8_t IMAGING_LED_PULSE = 15;   // Micro-seconds.  
const int IMAGING_CLICK_PAUSE = 7300;   // Micro-seconds.  
#endif

// For reading ADC to get temperature
const float TEMP_CONVERSION_FACTOR = 3.3f / (1 << 12);
const float ADC_OFFSET = 0.706;
const float ADC_DIVISOR = 0.001721;

// command inputs etc
const uint8_t MAX_INPUT_LEN = 8;
char command[10];
struct FocuserCommand buildCommand(char *text);

uint32_t locationRequired = 0;
uint32_t locationReported = 0;
uint16_t unusedStepDelay = 0;

//------------------------------
void motorControl();
void checkInput();
int getTemperature();
size_t appendChar(char *buffer, char c);
size_t clearBuffer(char *buffer);
void actionCommand(struct FocuserCommand c);
void initGPIOpins();
//==============================
int main(){
  stdio_init_all();
#ifdef PICOW
  cyw43_arch_init();
#endif
  initGPIOpins();

  // Configure ADC for temperature readings
  adc_init();
  adc_set_temp_sensor_enabled(true);
  adc_select_input(4);

  multicore_launch_core1(motorControl);
  checkInput();
  return 0; // but never
}

//===================================
void checkInput(){
  size_t cursor = 0;

  while (true){
    char ch = getchar_timeout_us(0);
    switch (ch){
    case ':':
      clearBuffer(command);
      break;
    case '#': // end of command
      actionCommand(buildCommand(command));
      break;
    default:
      if ((ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'Z'))
      {
        cursor = appendChar(command, ch);
      }
    }
    // Keep an eye on where the stepper is
    if (multicore_fifo_rvalid){
      multicore_fifo_pop_timeout_us(10, &locationReported);
    }

    sleep_ms(10);
  }
}

// append a character to the input buffer and return new length
size_t appendChar(char *buffer, char c){
  if (strlen(buffer) == MAX_INPUT_LEN)
    return MAX_INPUT_LEN;
  buffer[strlen(buffer)] = c;
  return strlen(buffer);
}

// clear contents of the specified buffer
size_t clearBuffer(char *buffer){
  memset(buffer, '\0', MAX_INPUT_LEN);
  return strlen(buffer);
}

//===================================
struct FocuserCommand buildCommand(char *text){

  struct FocuserCommand command;
  clearBuffer(command.action);
  command.value = 0;

  strncpy(command.action, text, 2);

  if (strlen(text) < 3){
    return command;
  }

  char numerics[5]; // Allocate space for substring
  int i;
  for (i = 2; i < 6; i++){ // Copy characters 2..5
    numerics[i - 2] = text[i];
  }
  numerics[4] = '\0'; // Manually add null terminator

  int temp = -1;
  int ix = 0;
  char commandList[] = "SN.SP.YT.YB.DB.SD";
  char thing[3];

  // Check if this command is in the list
  for (ix = 0; ix < 12; ix += 3){
    thing[0] = commandList[ix];
    thing[1] = commandList[ix + 1];
    thing[2] = '\0';
    if (strncmp(command.action, thing, 2) == 0){
      ix = 99;
      command.value = (uint32_t)strtol(numerics, NULL, 16);
    }
  }

  return command;
}

/*
Un-mentioned in the Moonlite protocol is the trick of doubling the temperature
value. The value is halved by the driver at the server end in order to get round
sending floating-point in binary.  Which of course works, only as long as the
resolution for measuring temperature is no better than 0.5 of a degree.
*/
int getTemperature(){
  uint16_t raw = adc_read();
  const float conversion_factor = 3.3f / (1 << 12);
  float result = raw * conversion_factor;
  float temp = 27 - (result - 0.706) / 0.001721;
  //printf("Temp = %f C\n", temp);
  return (int)temp * 2;
}

void initGPIOpins(){
  gpio_init(TMC_STEP);
  gpio_init(TMC_DIRECTION);
  gpio_init(TMC_STOPPED);

  gpio_init(LED_PIN);
  gpio_init(SW_DECREASE);
  gpio_init(SW_INCREASE);

  gpio_set_dir(TMC_STEP, GPIO_OUT);
  gpio_set_dir(TMC_DIRECTION, GPIO_OUT);
  gpio_set_dir(TMC_STOPPED, GPIO_OUT);

  gpio_set_dir(LED_PIN, GPIO_OUT);
  gpio_set_dir(SW_DECREASE, GPIO_IN);
  gpio_set_dir(SW_INCREASE, GPIO_IN);

  gpio_pull_up(SW_DECREASE);
  gpio_pull_up(SW_INCREASE);

#ifdef IMAGING
  gpio_init(SW_IMAGING);
  gpio_init(LED_IMAGING);
  gpio_set_dir(LED_IMAGING, GPIO_OUT);
  gpio_set_dir(SW_IMAGING, GPIO_IN);
  gpio_pull_up(SW_IMAGING);
#endif
}
