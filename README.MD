Tester for driving the LED displays on a GamePlan pinball machine

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