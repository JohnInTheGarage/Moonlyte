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

	moving          bool
	debugging       bool
	locationCurrent int32
	locationTarget  int32
	activeTarget    bool
	stepDelay       int16
)

const (
	maxSteps     int32         = 20
	stepDuration time.Duration = 2 * time.Millisecond
	locationMin  int32         = 0
	locationMax  int32         = 30000
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

		// activeTarget is turned on by :FG# command (or manual input)
		if activeTarget {
			if locationTarget != locationCurrent {
				moving = true
				doSomeStepping()
			}
		}

	}

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

	distance = locationTarget - locationCurrent
	debug(fmt.Sprintln("Location ", locationCurrent, ", distance :", distance))

	if distance == 0 {
		debug("calling halt as distance = 0")
		haltStepper()
	}

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

// =============================
func boj() {

	time.Sleep(1 * time.Second)
	uart.Configure(machine.UARTConfig{TX: tx, RX: rx})
	tmcStep.Configure(machine.PinConfig{Mode: machine.PinOutput})
	tmcDirection.Configure(machine.PinConfig{Mode: machine.PinOutput})
	tmcEnable.Configure(machine.PinConfig{Mode: machine.PinOutput})
	tmcDirection.Low()
	swIncrease.Configure(machine.PinConfig{
		Mode: machine.PinInputPullup,
	})
	swDecrease.Configure(machine.PinConfig{
		Mode: machine.PinInputPullup,
	})
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
		haltStepper()
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
 * Handle two optional switches to allow manual adjustment of
 * focus increase or decrease.  Added so that I can focus when
 * the non-astro camera is connected.  That requires me to watch
 * the camera screen while at the scope instead of having the
 * computer do it remotely.
 * This also means we have to prevent going below zero, which
 * the Indigo/Ascom/Indi software normally handles.
 */
func checkSwitches() {
	if swIncrease.Get() == false {
		locationTarget = locationCurrent + 5
	}

	if swDecrease.Get() == false {
		locationTarget = locationCurrent - 5
	}

	// if both buttons are pressed - no movement
	if locationTarget == locationCurrent {
		return
	}

	time.Sleep(10 * time.Millisecond) // otherwise it operates very quickly
	if locationTarget < locationMin {
		locationTarget = locationMin
	}
	if locationTarget > locationMax {
		locationTarget = locationMax
	}
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
		enableStepper()
		moving = true
	}
}

// ======================// ======================
func haltStepper() {
	moving = false
	tmcEnable.High()
	debug("Halted stepper")
}

// ======================
func enableStepper() {
	tmcEnable.Low()
	debug("Stepper enabled")
}

// ======================
func debug(msg string) {
	if debugging {
		println(msg)
	}
}
