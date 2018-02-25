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
Sounds for Bossy:
0 = High tone (maybe a puck bouncing ? ?) - maybe have this when you hit a lighted puck
1 = Star spangled banner - after warmup period
2 = up then down tone - like crowd doing the wave? (maybe on ball launch have this?)
3 = icing / asteroids fire -- pop bumpers and sling shots (defense)
4 = reset
5 = whistle, up and down tone, and asteroids icing (end of game)
6 = n/a
7 = whistle only - ball drain, and 2 for ball launch?
8 = low tones (bouncing ball..) - outlanes?
9 = high tones (bouncing ball) ?? - when we are moving the lit puck to a new place
10 = nothing
11 = puck bounce (10 pt shot)
12 = ra ra , ra-ra-ra, ra ra, ra-ra-ra (when you set up for a line shot)
13 = charge - (add player / credit)
14 = nothing
15 = nothing



From https://github.com/google/periph/blob/master/host/rpi/rpi.go

P1_13 gpio.PinIO = bcm283x.GPIO27 // Low,  <<--Data In - pin 14 on lsb 595
P1_14 pin.Pin    = pin.GROUND     //
P1_15 gpio.PinIO = bcm283x.GPIO22 // Low, <<--Clock Data - pin 11 on '595s
P1_16 gpio.PinIO = bcm283x.GPIO23 // Low, <<--Latch Data - pin 12 on '595s

*/

import (
	"flag"
	"fmt"
	"log"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host"
	"periph.io/x/periph/host/rpi"
)

var _disp [5][7]byte //this holds what we want to show on the display. Bytes are in terms of what the 74ls48 supports (0x0f is blank)
var _sound byte

const (
	blank           byte = 0x0f //what is sent to the 7448 on the display board to blank the 7 seg disp
	blankScore           = -1   //the numeric number that can be passed in as an integer to clear the display
	creditMatchDisp      = 4    //the number in the display array for the credit display
	creditLSD            = 6    //position in the display array for the 1's credit disp digit
	creditMSD            = 0    //position in the display array for the 10's credit disp digit

	pinDataClk  = 15
	pinLatchClk = 16
)

//const blank byte = 0x0f
//const blankScore = -1
//const creditMatchDisp = 4

var endLoop bool

func main() {
	clearDisplays()
	var play int

	flag.IntVar(&play, "snd", 0x0f, "Pass the number of the sound to play")
	flag.Parse()

	_sound = byte(play)

	if !rpi.Present() {
		fmt.Println("Not running on a raspberry pi. Debug information is displayed only")
		mainDebug()
		return
	}

	mainRPI()
}

func mainDebug() {
	printDisplays()

	fmt.Println("Setting player 1 = 1234, player 2 = 7654321, player 3 = 1234567, player 4 = 9080706, match = 8877 ")
	// Load all the drivers:
	setScore(0, 0)
	setScore(1, 10)
	setScore(2, 100)
	setScore(3, 1000)

	printDisplays()

	fmt.Println("Blanking Credits and Balls")
	setCredits(blankScore)
	setBallInPlay(blankScore)
	printDisplays()

	fmt.Println("Setting Ball in Play 3, Credits 56")
	setCredits(56)
	setBallInPlay(3)

	printDisplays()
}
func mainRPI() {
	endLoop = false
	fmt.Println("Running on rpi")

	if _sound != 0x0f {
		fmt.Printf("Playing sound %d", _sound)
	} else {
		fmt.Println("Not playing sound")
	}

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}
	initPorts()

	go runDisplays()

	//throwing out some test data
	setScore(0, 7654321)
	setScore(1, 7654321)
	setScore(2, 7654321)
	setScore(3, 7654321)

	setCredits(21)
	setBallInPlay(43)

	printDisplays()

	time.Sleep(2000 * time.Millisecond) //wait 2 seconds before turning off.
	if _sound != 0x0f {

		_sound = 0x0f
		endLoop = true
		time.Sleep(20 * time.Millisecond)
		return
	}

	time.Sleep(1 * time.Second)
	clearDisplays()
	_sound = 0x0f
	endLoop = true
	time.Sleep(500 * time.Millisecond)
}

func initPorts() {
	rpi.P1_13.Out(gpio.Low)
	rpi.P1_15.Out(gpio.Low)
	rpi.P1_16.Out(gpio.Low)

}

func clearDisplays() {
	for i := 0; i < len(_disp); i++ {
		blankDisplay(i)
	}
}

//dspOut, sends all of the bytes out for controlling the displays
//Data needs to be first followed by clock followed by digits
//MSB needs to be sent first as well
func dspOut(digits byte, clock byte, dspData byte, sndData byte) {
	thirdReg := sndData<<4 | dspData&0x0f //no need to mask the dsp data really, but just in case
	shiftOut(thirdReg)
	shiftOut(clock)
	shiftOut(digits)
	pulse(pinLatchClk) //latch output of shift registers
}

//shifOut sends value "val" passed in to the '595 and latches the output
func shiftOut(val byte) {

	var a byte

	a = 0x80 //msb first

	for b := 1; b <= 8; b++ {
		if val&a > 0 {
			rpi.P1_13.Out(gpio.High)
		} else {
			rpi.P1_13.Out(gpio.Low)
		}

		pulse(pinDataClk) //pulse clock line
		a >>= 1
	}
}

func pulse(pin int) {
	switch pin {
	case pinDataClk:
		rpi.P1_15.Out(gpio.High)
	case pinLatchClk:
		rpi.P1_16.Out(gpio.High)
	}

	//delay here
	//time.Sleep(1 * time.Microsecond) //this should be enough for a HC595 I think?

	switch pin {
	case pinDataClk:
		rpi.P1_15.Out(gpio.Low)
	case pinLatchClk:
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

	for {
		digitStrobe = 0x01
		//data = 0x06

		for digit = 0; digit < 7; digit++ {
			var clkOut byte = 0x01
			for display = 0; display < 4; display++ {
				data = _disp[display][digit]

				dspOut(0, clkOut, data, _sound)
				clkOut <<= 1
			}

			//strobing the digit here, which is why we took it out of the other for loop
			data = _disp[creditMatchDisp][digit]
			dspOut(digitStrobe, clkOut, data, _sound)
			digitStrobe <<= 1                  //shifting over for the next digit
			time.Sleep(100 * time.Microsecond) //230 ms should be 120hz to the displays?
		}

		if endLoop {
			break
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

func numToArray(number int32) ([]byte, error) {
	var scoreArr []byte

	var remainder int32
	tmpScore := number

	for {
		remainder = tmpScore % 10
		scoreArr = append(scoreArr, byte(remainder))
		tmpScore /= 10

		if tmpScore == 0 {
			break
		}
	}

	return scoreArr, nil
}

//assumption is 7 digit display, so we will blank all remaining digits the score passed in didn't set
func setScore(dispNum int, score int32) {
	scoreArr, _ := numToArray(score)

	_disp[dispNum] = [...]byte{blank, blank, blank, blank, blank, blank, blank} //initialize to blank disp

	//copy the score into the display array
	for i, num := range scoreArr {
		_disp[dispNum][i] = num
	}
}

//pretty sure match and ball in play are the same display (digits 4 and 3), Credit is 0 and 6
func setBallInPlay(ball int32) {
	ballDisp := _disp[creditMatchDisp][3:5]
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
		ballDisp[0] = ballArr[0]
		ballDisp[1] = blank
	}
}

//for some reason GamePlan uses digit 6 and 0
func setCredits(credit int32) {

	if credit == blankScore {
		_disp[creditMatchDisp][creditMSD] = blank
		_disp[creditMatchDisp][creditLSD] = blank
		return
	}

	creditArr, _ := numToArray(credit)

	if len(creditArr) == 2 {
		_disp[creditMatchDisp][creditLSD] = creditArr[0]
		_disp[creditMatchDisp][creditMSD] = creditArr[1]
	} else {
		_disp[creditMatchDisp][creditLSD] = creditArr[0]
		_disp[creditMatchDisp][creditMSD] = blank
	}
}

//loops all digits through the displays
func dispDiagnostics() {

	clearDisplays()

	var cnt int32
	cnt = 1111111

	for i := 1; i < 10; i++ {
		for dsp := 1; dsp < 5; dsp++ {
			setScore(dsp, cnt)
		}
		cnt += 1111111
		time.Sleep(1000 * time.Millisecond)
	}
}

func printDisplays() {
	for i, val := range _disp {
		fmt.Printf("Display Array %d: ", i)
		fmt.Println(val)
	}

	fmt.Println("Displays as shown:")
	for i, d := range _disp {
		fmt.Printf("Disp #%d: ", i+1)
		for digit := len(d) - 1; digit >= 0; digit-- {
			fmt.Printf("%d ", d[digit])
		}

		fmt.Print("\n")
	}
}
