package main

/*Tester for driving the LED displays on a GamePlan pinball machine

The displays are LED "ttl" driven (has a 7417 in the original schematic, with a clamping diode on the output)

What needs to be driven:

For Display Data:
Bit0
Bit1
Bit2
Bit3

For Digit Control:
Digit 0
Digit 1
Digit 2
Digit 3
Digit 4
Digit 5
Digit 6

For which Display module (rising edge locks in data):
Clk1
Clk2
Clk3
Clk4
Clk5

Overall Display Control:
Enable (Acitve Low)


Overall:
Since all same digits are tied together on enable (Digit0 on all boards are wired together), we need to:

Set Enable = Low
For each digit:
	For each display:
		Load Data on Bit0-Bit3
		Pulse Clck data line for display
	Pulse Digit line (low to high)



Alowed Digits (based on 74ls48 datasheet):
0-9 shows the respective digit
10 = [ (really a c, the lower part of the display)
11 = ] (backwards c, the lower part of the display)
12 = u (on the top of the display)
13 = c +  _ (top display c with a _ at the bottom)
14 = t
15 = blank (nothing displayed)



Use of serial buffering:
74HC595: Pin 14 = serial data in
Pin 11 = shcp clock data in (every serial bit in, latches into the register )
Pin 12 = stcp clock data out (latch on the outputs )

We are going to daisy chain the 595s, so we send 24 bits every time before stcp / latch out


74HC595 for digit control (first in chain - pin 4 on rasp pi to pin 14 on 595)
74HC595 for clock control (second in chain - pin 9 to pin 14)
74HC595 for data control (third in chain - pin 9 to pin 14)


3333 3333 2222 2222 1111 1111
SND- DATA	 CLOCK-	 -DIGITS-


SND = Sound board control (when ready)


From https://github.com/google/periph/blob/master/host/rpi/rpi.go

P1_13 gpio.PinIO = bcm283x.GPIO27 // Low,  <<--Data In - pin 14 on lsb 595
P1_14 pin.Pin    = pin.GROUND     //
P1_15 gpio.PinIO = bcm283x.GPIO22 // Low, <<--Clock Data - pin 11 on '595s
P1_16 gpio.PinIO = bcm283x.GPIO23 // Low, <<--Latch Data - pin 12 on '595s

*/

import (
	"fmt"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host/rpi"
)

var _disp [5][7]byte //this holds what we want to show on the display. Bytes are in terms of what the 74ls48 supports (0x0f is blank)
const blank byte = 0x0f
const blankScore = -1
const creditMatchDisp = 4

func main() {
	initDisplays()
	printDisplays()

	fmt.Println("Setting player 1 = 1234, player 2 = 7654321, player 3 = 1234567, player 4 = 9080706, match = 8877 ")
	// Load all the drivers:
	setScore(0, 1234)
	setScore(1, 7654321)
	setScore(2, 1234567)
	setScore(3, 9080706)
	setScore(4, 8877)

	printDisplays()

	fmt.Println("Blanking Credits and Balls")
	setCredits(blankScore)
	setBallInPlay(blankScore)
	printDisplays()

	fmt.Println("Setting Ball in Play 3, Credits 56")
	setCredits(56)
	setBallInPlay(3)

	printDisplays()
	/*if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	//go runDisplays()
	*/
	//test the displays out

	/*for l := gpio.Low; ; l = !l {
		// Lookup a pin by its location on the board:
		if err := rpi.P1_33.Out(l); err != nil {
			log.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
	}*/
}

func initPorts() {
	rpi.P1_13.Out(gpio.Low)
	rpi.P1_15.Out(gpio.Low)
	rpi.P1_16.Out(gpio.Low)

}

func initDisplays() {
	for i := 0; i < len(_disp); i++ {
		blankDisplay(i)
	}
}

//dspOut, sends all of the bytes out for controlling the displays
//Data needs to be first followed by clock followed by digits
//MSB needs to be sent first as well
func dspOut(digits byte, clock byte, data byte) {
	shiftOut(data)
	shiftOut(clock)
	shiftOut(digits)
	pulse(16)
}

//shifOut sends value "val" passed in to the '595 and latches the output
func shiftOut(val byte) {

	var a byte

	a = 0x80 //msb first

	for b := 1; b < 8; b++ {
		if val&a > 0 {
			rpi.P1_13.Out(gpio.High)
		} else {
			rpi.P1_13.Out(gpio.Low)
		}

		pulse(15)
		a >>= 1
	}
}

func pulse(pin int) {
	switch pin {
	case 15:
		rpi.P1_15.Out(gpio.High)
	case 16:
		rpi.P1_16.Out(gpio.High)
	}

	//delay here
	time.Sleep(1 * time.Microsecond) //this should be enough for a HC595 I think?

	switch pin {
	case 15:
		rpi.P1_15.Out(gpio.Low)
	case 16:
		rpi.P1_16.Out(gpio.Low)
	}
}

/*
For each digit:
	For each display:
		Load Data on Bit0-Bit3
		Pulse Clck data line for display
	Pulse Digit line (low to high)
*/
func runDisplays() {

	var digit, display, data, digitStrobe byte
	digitStrobe = 0

	for {
		for digit = 0; digit < 7; digit++ {
			for display = 1; display < 4; display++ {
				data = _disp[display][digit]
				dspOut(digitStrobe, display, data)
			}

			//strobing the digit here, which is why we took it out of the other for loop
			digitStrobe = digit
			data = _disp[4][digit]
			dspOut(digitStrobe, display, data)

			time.Sleep(230 * time.Microsecond) //230 ms should be 120hz to the displays?
		}

		//loop forever
	}
}

func setDisplay(dispNum int, digits []byte) {
	for i, d := range digits {
		_disp[dispNum][i] = d
	}
}

func blankDisplay(dispNum int) {
	_disp[dispNum] = [...]byte{blank, blank, blank, blank, blank, blank, blank} //initialize to blank disp
}

//assumption is 7 digit display, so we will blank all remaining digits the score passed in didn't set
func setScore(dispNum int, score int) {
	scoreArr, _ := numToArray(score)

	_disp[dispNum] = [...]byte{blank, blank, blank, blank, blank, blank, blank} //initialize to blank disp

	//copy the score into the display array
	for i, num := range scoreArr {
		_disp[dispNum][len(_disp[dispNum])-len(scoreArr)+i] = num
	}
}

func numToArray(number int) ([]byte, error) {
	var scoreArr []byte

	var remainder int
	tmpScore := number

	for {
		remainder = tmpScore % 10
		scoreArr = append([]byte{byte(remainder)}, scoreArr...)
		tmpScore /= 10

		if tmpScore == 0 {
			break
		}
	}

	return scoreArr, nil
}

func printDisplays() {
	for i, val := range _disp {
		fmt.Printf("Display %d: ", i)
		fmt.Println(val)
	}
}

//pretty sure match and ball in play are the same display (digits 1 and 2), Credit is 5 and 6
func setBallInPlay(ball int) {
	ballDisp := _disp[creditMatchDisp][5:7]
	if ball == blankScore {
		ballDisp[0] = blank
		ballDisp[1] = blank
		return
	}

	ballArr, _ := numToArray(ball)

	if len(ballArr) == 2 {
		ballDisp[0] = ballArr[0]
		ballDisp[1] = ballArr[1]
	} else {
		ballDisp[1] = ballArr[0]
		ballDisp[0] = blank
	}
}

func setCredits(credit int) {
	creditDisp := _disp[creditMatchDisp][1:3]

	if credit == blankScore {
		creditDisp[0] = blank
		creditDisp[1] = blank
		return
	}

	creditArr, _ := numToArray(credit)

	if len(creditArr) == 2 {
		creditDisp[0] = creditArr[0]
		creditDisp[1] = creditArr[1]
	} else {
		creditDisp[0] = creditArr[0]
		creditDisp[1] = blank
	}
}
