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
15 = blank

Digits that are stored in the array are Least Siginificant Digit to Most Significant Digit
Thus array[0] = 1's digit, array[1] = 10's digit etc.

