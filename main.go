package main

/*
* Emulates the moonlite protocol for Auto-focusing telescopes on a Raspberry PICO using Tinygo.
* My implementation uses a TMC2009 driver circuit to provide more power than direct connection
* between Stepper and GPIO pins.
*
* Also, a couple of non-standard additions : DBnnnn to switch debugging on and off, plus
* the full/half step option is hijacked to allow reversing the stepper past the place
* where it finds itself at power-on.  This is used to fix the situation where I forget to
* return the focus to zero before power-off (No non-volatile memory on Pico to store actual location).

* Its not quite in the shape I envisaged.  It seemed ideal to devote one core to driving the stepper
* and the other to comms with the server, but in tinygo scheduling is cooperative not pre-emptive and
* only uses one core.
* Therefore the stepping activity is broken down into small groups of steps mixed in with the serial comms
* with the server, to allow it to apparently be happening at the same time.
*
* 2023-10-31 Added manual increase/decrease buttons.
*
* 2023-12-09 Reversed direction of motor for Version 2 hardware change - Motor inverted from V1
*            to use belt drive instead of direct connection to the focuser shaft.
*
* 2023-12-15 Reduce "chugging" on manual inputs (caused by enabling and disabling the stepper
*            for the short bursts used during manual-focus button-presses).
*
* 2024-01-20 Add a switch for activating an Intervalometer.  My Canon EOS-M camera lacks that feature,
*            and my commercial Intervalometer only goes to 40 (secs or mins) and when it does e.g.
*            30 seconds, its on for 30 and then off for 30.  I need on for 30 then off for a few
* 		     seconds so the camera can start another exposure without pointless waiting.
* 			 If you want to exclude this feature, remove lines that include "imaging"
 */

// flashing : tinygo flash -target=pico main.go

// Serial code copied from https://github.com/tinygo-org/tinygo/blob/release/src/examples/echo/echo.go

import (
	"fmt"
	"machine"
	"strconv"
	"time"
)

type (
	FocuserCommand struct {
		action string
		value  int32
	}
)

var (
	versionString = "10"
	uart          = machine.Serial
	tx            = machine.UART_TX_PIN
	rx            = machine.UART_RX_PIN
	tmcStep       = machine.GP16
	// Setting this Low turns the stepper clockwise, High anti-clockwise
	tmcDirection = machine.GP17
	tmcEnable    = machine.GP18
	//two optional switches for manual adjustments
	swIncrease = machine.GP20
	swDecrease = machine.GP19
	// Intervalometer
	swImaging  = machine.GP10
	imagingLED = machine.GP11

	moving          bool
	debugging       bool
	locationCurrent int32
	locationTarget  int32
	activeTarget    bool
	stepDelay       int16
	stepperEnabled  bool
	stepperCooling  time.Time // A delay used after reaching the target location, but before disabling the stepper
	imagingEnabled  bool
	imagingActive   bool
	imagingClose    time.Time // Close the shutter when this time is reached.
	imagingOpen     time.Time // Open the shutter when this time is reached.
)

const (
	maxSteps     int32         = 20
	stepDuration time.Duration = 2 * time.Millisecond
	locationMin  int32         = 0
	locationMax  int32         = 30000
	// Switches are on when the pin's .Get() method returns false
	sw_on           bool          = false
	imagingSeconds  time.Duration = 60 * time.Second // exposure time
	imagingPause    time.Duration = 8 * time.Second  // Allow time to save RAW image
	imagingLEDPulse time.Duration = 11 * time.Microsecond
)

// ======================
func main() {
	var inbyte byte

	boj()

	data := make([]byte, 0)
	for {

		for uart.Buffered() > 0 {
			inbyte, _ = uart.ReadByte()
			switch inbyte {
			case ':':
				{
					data = make([]byte, 0)
				}
			case '#':
				{
					input := string(data[:])
					actionCommand(buildCommand(input))
					break
				}
			default:
				{
					data = append(data, inbyte)
					continue
				}
			}
		}

		checkSwitches() // any manual input?

		// If you turn on imaging while the server is adjusting focus, expect disappointment!
		if imagingEnabled || imagingActive {
			// Imaging could be switched off while shutter is open
			handleImaging()
		}

		// activeTarget is turned on by :FG# command (or manual input)
		if activeTarget {
			if locationTarget != locationCurrent {
				moving = true
				doSomeStepping()
			}
			if time.Now().After(stepperCooling) {
				debug(fmt.Sprintln("Now ", time.Now().Nanosecond(), ", cooling :", stepperCooling.Nanosecond()))
				haltStepper()
				activeTarget = false
			}
		}

	} // end of "for ever"

}

/*
* Stepping is a blocking activity so only do a few at a time
* Otherwise we can't respond to Indigo quickly enough
 */

func doSomeStepping() {
	var distance = locationTarget - locationCurrent
	var steps = distance
	var delta int32

	if distance < 0 {
		delta = -1
		steps *= delta // to get a positive number
		tmcDirection.Low()
	} else {
		tmcDirection.High()
		delta = 1
	}
	if steps > maxSteps {
		steps = maxSteps
	}

	for steps > 0 {
		tmcStep.High()
		time.Sleep(stepDuration)
		tmcStep.Low()
		time.Sleep(stepDuration)
		locationCurrent += delta
		steps--
	}
	stepperCooling = time.Now().Add(time.Second * 3)

	distance = locationTarget - locationCurrent
	if distance == 0 {
		moving = false
	}
	debug(fmt.Sprintln("Location ", locationCurrent, ", distance :", distance))
}

// ======================
/* Respond to INDI / ASCOM / INDIGO controlling software
 */
func respond(msg string) {
	// no CR for real use >>> uart.Write([]byte(msg + "#\r\n"))
	uart.Write([]byte(msg + "#"))
}

// ======================
func buildCommand(text string) FocuserCommand {

	var value int64
	var err error
	var command FocuserCommand

	if len(text) < 2 {
		command.action = text[0:1]
		command.value = 0
		return command
	}
	if len(text) < 3 {
		command.action = text[0:2]
		command.value = 0
		return command
	}

	if text[0:2] == "SN" || text[0:2] == "SP" || text[0:2] == "YT" || text[0:2] == "YB" || text[0:2] == "DB" || text[0:2] == "SD" {
		value, err = strconv.ParseInt(text[2:], 16, 0)
	}
	if err != nil {
		fmt.Println(err)
	}

	command.action = text[0:2]
	command.value = int32(value)
	debug(fmt.Sprintln("command :", command.action, ",", command.value))

	return command
}

// =============================
func floatAsHex(fnum float32) string {
	return fmt.Sprintf("%04X", fnum)
}

// =============================
func intAsHex(num int32) string {
	return fmt.Sprintf("%04X", num)
}

// =============================
func int16AsHex(num int16) string {
	return fmt.Sprintf("%02X", num)
}

// =============================
func int8AsBinary(num int8) string {
	return fmt.Sprintf("%b", num)
}

// =============================
/*
machine.ReadTemperature() returns milli Celcius, so divide by 1000,
But also check the send temp for moonlite nonsense
*/
func ReadTemp() int32 {
	temp := machine.ReadTemperature() / 1000
	return temp
}

/*
* set shutter Open time to 5 seconds after shutter Close time, to allow camera to
* save the RAW image - which takes a short while.
 */
func handleImaging() {
	if time.Now().After(imagingOpen) {
		debug("imaging:Open shutter")
		fireImagingShutterLED()
		imagingClose = time.Now().Add(imagingSeconds)
		imagingOpen = imagingClose.Add(imagingPause)
		imagingActive = true
	}

	if time.Now().After(imagingClose) {
		if imagingActive {
			debug("imaging:Close shutter")
			fireImagingShutterLED()
		}
		imagingActive = false
	}
}

// =============================
/*
* Idea lifted from https://github.com/MATT-ER-HORN/multiCameraIrControl
* (Canon cameras only, others are available)
* original work (apparently) by Sebastien Setz which seems to be unavailable now
 */
func fireImagingShutterLED() {
	for i := 0; i < 16; i++ {
		imagingLED.High()
		time.Sleep(imagingLEDPulse)
		imagingLED.Low()
		time.Sleep(imagingLEDPulse)
	}

	time.Sleep(7330 * time.Microsecond)

	for i := 0; i < 16; i++ {
		imagingLED.High()
		time.Sleep(imagingLEDPulse)
		imagingLED.Low()
		time.Sleep(imagingLEDPulse)
	}
}

// =============================
func boj() {
	time.Sleep(5 * time.Second)
	uart.Configure(machine.UARTConfig{TX: tx, RX: rx})
	tmcStep.Configure(machine.PinConfig{Mode: machine.PinOutput})
	tmcDirection.Configure(machine.PinConfig{Mode: machine.PinOutput})
	tmcEnable.Configure(machine.PinConfig{Mode: machine.PinOutput})
	tmcDirection.Low()
	swIncrease.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	swDecrease.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	swImaging.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	imagingLED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	imagingOpen = time.Now()
	imagingClose = imagingOpen
	haltStepper()
}

//================================================
//================================================

/* ======================
* Ignored commands :
* C -> No delay expected in fetching Temp, GT can do it
* GT, SC, +, -, Y+, Y-, PO, ZX -> Temperature coefficient focusing not used
* YM
* SP, YT -> not allowing driver to set potentially wrong values
 */
func actionCommand(command FocuserCommand) {

	switch command.action {
	case "DB": // >>>>>>>>>>>>>>>> non-standard!  May be removed
		setDebug(command.value)
	case "FQ":
		emergencyStop()
	case "FG":
		goToTarget()
	case "GH":
		sendHalfstepStatus()
	case "GC":
		sendTempCoefficient()
	case "GI":
		sendMovingStatus()
	case "GN":
		sendTargetPosition()
	case "GP":
		sendCurrentPosition()
	case "GT":
		sendTemperature()
	case "GV":
		sendVersion()
	case "GD":
		sendStepDelay()
	case "SD":
		setStepDelay(command.value)
	case "SF":
		hijackFullStepMode()
	case "SH":
		hijackHalfStepMode()
	case "SN":
		setNewTarget(command.value)
	case "YB":
		setBacklash()
	case "ZB":
		sendBacklash()
	case "ZT":
		sendMaxSteps()
	case "ZA":
		sendAverageTemp()
	default:
		{
		}
	}

}

/*
 * Handle three optional switches; 2 to allow manual adjustment of
 * focus increase or decrease, and 1 to activate the Intervalometer.
 * Added so that I can use my old mirrorless non-astro camera.
 * That requires me to watch the camera screen while at the scope
 * instead of having the computer do it remotely.
 * === Intervalometer and manual focus are mutually exclusive. ===
 * This also means we have to prevent focus going below zero, which
 * the Indigo/Ascom/Indi software normally handles.
 */
func checkSwitches() {

	if swImaging.Get() == sw_on {
		imagingEnabled = true
		return
	}
	imagingEnabled = false

	// if both buttons are pressed or not pressed - no movement
	if swIncrease.Get() == swDecrease.Get() {
		return
	}

	if swIncrease.Get() == sw_on {
		locationTarget = locationCurrent + 5
	}

	if swDecrease.Get() == sw_on {
		locationTarget = locationCurrent - 5
	}

	debug(fmt.Sprintln("checkswitches [", locationCurrent, "] to [", locationTarget, "]"))

	time.Sleep(10 * time.Millisecond) // otherwise it operates very quickly
	if locationTarget < locationMin {
		locationTarget = locationMin
	}
	if locationTarget > locationMax {
		locationTarget = locationMax
	}
	debug("checkswitches calls gotoTarget")
	goToTarget()

}

// ======================
func setDebug(value int32) {
	debugging = (value > 0)
	if debugging {
		debug("debugging on")
	} else {
		debug("debugging off")
	}
}

// ======================
func sendAverageTemp() {
}

// ======================
func sendMaxSteps() {
}

// ======================
func sendBacklash() {
}

// ======================
func setBacklash() {
}

// ======================
func setNewTarget(value int32) {
	activeTarget = false
	locationTarget = value
}

// ======================
/*
* Since the TMC2209 is already micro-stepping,
* Full step and Half Step commands are re-purposed
* to overcome lack of known start-point
 */
func hijackFullStepMode() {
	// Restore normal direction
	tmcDirection.High()
}

// ======================
/*
* Reset the "zero" position.  Go backwards 1000 steps and
* change location to zero
* The User is expected to be monitoring progress carefully!
 */
func hijackHalfStepMode() {
	locationCurrent = 1000
	locationTarget = 0
	tmcDirection.Low()
	goToTarget()
}

// =======================
func sendStepDelay() {
	respond(int16AsHex(stepDelay))
}

// =======================
func setStepDelay(value int32) {
	stepDelay = int16(value)
}

// ======================
func sendVersion() {
	respond(versionString)
}

// =========================
func sendTempCoefficient() {
	respond("00")
}

// ======================
/*
Un-mentioned in the Moonlite protocol is the trick of doubling the temperature value.
The value is halved by the driver at the server end in order to get round sending
floating-point in binary.  Which of course works, only as long as the resolution for
measuring temperature is no better than 0.5 of a degree.
*/
func sendTemperature() {
	respond(intAsHex(ReadTemp() * 2))
}

// ======================
func sendCurrentPosition() {
	respond(intAsHex(locationCurrent))
}

// ======================
func sendTargetPosition() {
	respond(intAsHex(locationTarget))
}

// ======================
func sendMovingStatus() {
	if moving {
		respond("01")
	} else {
		respond("00")
	}
}

// ======================
func sendHalfstepStatus() {
	respond("00")
}

// ======================
func goToTarget() {
	activeTarget = true
	debug(fmt.Sprintln("Moving from [", locationCurrent, "] to [", locationTarget, "]"))
	if locationCurrent != locationTarget {
		moving = true
	}
	if stepperEnabled == false {
		debug(" tmcEnable is not enabled ")
		enableStepper()
	}
}

// ======================
func emergencyStop() {
	moving = false
	activeTarget = false
	locationTarget = locationCurrent
	haltStepper()
	debug("emergency stop")
}

// ======================
func haltStepper() {
	tmcEnable.High()
	stepperEnabled = false
	debug("Halted stepper")
}

// ======================
func enableStepper() {
	tmcEnable.Low()
	stepperEnabled = true
	debug("Stepper enabled")

}

// ======================
func debug(msg string) {
	if debugging {
		println(msg)
	}
}
