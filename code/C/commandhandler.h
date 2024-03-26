/*  Testing sleep timings
extern uint64_t a;
extern uint64_t b;
extern uint64_t c;
extern uint64_t d;
extern uint64_t e;
extern uint64_t f;
extern uint64_t hw_h2;
extern uint64_t hw_h1;
*/

extern uint32_t locationRequired;
extern uint32_t locationReported;
extern uint16_t unusedStepDelay;

extern const uint32_t SIGNAL_STOP;
extern const uint32_t SIGNAL_GO;
extern const uint32_t SIGNAL_BACK1000;

extern int getTemperature();

struct FocuserCommand
{
  char action[3];
  uint32_t value;
};

void sendAverageTemp();
void sendMaxSteps();
void sendBacklash();
void setBacklash();
void setNewTarget(int value);
void hijackFullStepMode();
void hijackHalfStepMode();
void sendStepDelay();
void setStepDelay(int value);
void sendVersion();
void sendTempCoefficient();
void sendTemperature();
void sendCurrentPosition();
void sendTargetPosition();
void sendMovingStatus();
void sendHalfstepStatus();
void goToTarget();
void emergencyStop();
void respond(char *text);
void respondAsHex2(int num);
void respondAsHex4(int num);
void passToCore1(int item);
