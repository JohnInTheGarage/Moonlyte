
extern const uint32_t SIGNAL_STOP;
extern const uint32_t SIGNAL_GO;
extern const uint32_t SIGNAL_BACK1000;
extern const uint8_t TMC_STEP;
extern const uint8_t TMC_STOPPED;
extern const uint8_t TMC_DIRECTION;
extern const uint8_t SW_IMAGING;
extern const uint8_t SW_DECREASE;
extern const uint8_t SW_INCREASE;
extern const uint8_t LED_IMAGING;
extern const uint8_t LED_PIN;
extern const uint8_t IMAGING_SECONDS;
extern const uint8_t IMAGING_PAUSE;
extern const uint8_t IMAGING_LED_PULSE;
extern const int IMAGING_CLICK_PAUSE;

uint32_t locationTarget = 0;
uint32_t locationCurrent = 0;
uint32_t newInput = 0;
int minStepDuration = 2; // milliseconds
int maxStepDuration = 40; // milliseconds
int stepDuration = 0;
int delta = 0;
int maxSteps = 20;
int stepsReqd = 0;

bool stepping;
bool imagingStarted;
bool shutterClickRequired;
bool shutterOpen;

// false = low, i.e. grounded, i.e. switched on (as pull-ups in use)
bool sw_on  = false; 
bool sw_off = true;

void handleImaging();
void handleStepping();
void handleNewInput();
void checkImagingSwitch(); 
void fireImagingShutterLED();
void allowManualFocus();
int checkStepperCooldown(int counter);
void setStepsDirection();
// these are called by alarm timer
int64_t triggerShutter(alarm_id_t id, void *user_data);
int64_t disableStepper(alarm_id_t id, void *user_data);