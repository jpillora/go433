package go433

/*
  Ported from:
  https://github.com/sui77/rc-switch/blob/master/RCSwitch.cpp

  RCSwitch - Arduino libary for remote control outlet switches
  Copyright (c) 2011 Suat Özgür.  All right reserved.

  Contributors:
  - Andre Koehler / info(at)tomate-online(dot)de
  - Gordeev Andrey Vladimirovich / gordeev(at)openpyro(dot)com
  - Skineffect / http://forum.ardumote.com/viewtopic.php?f=2&t=46
  - Dominik Fischer / dom_fischer(at)web(dot)de
  - Frank Oltmanns / <first name>.<last name>(at)gmail(dot)com
  - Andreas Steinel / A.<lastname>(at)gmail(dot)com

  Project home: http://code.google.com/p/rc-switch/
  This library is free software; you can redistribute it and/or
  modify it under the terms of the GNU Lesser General Public
  License as published by the Free Software Foundation; either
  version 2.1 of the License, or (at your option) any later version.
  This library is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
  Lesser General Public License for more details.
  You should have received a copy of the GNU Lesser General Public
  License awith int this library; if not, write to the Free Software
  Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
*/

import (
	"fmt"
	"time"

	"github.com/kidoman/embd"

	// _ "github.com/kidoman/embd/host/rpi"
)

// separationLimit: minimum microseconds between received codes, closer codes are ignored.
// according to discussion on issue #14 it might be more suitable to set the separation
// limit to the same time as the 'low' part of the sync signal for the current protocol.
const nSeparationLimit = 4600
const RCSWITCH_MAX_CHANGES = 1000

type RCSwitch struct {
	nProtocol          int
	nPulseLength       int
	nReceivedValue     int
	nRepeatTransmit    int
	nReceivedBitlength int
	nReceivedDelay     int
	nReceivedProtocol  int
	nReceiveTolerance  int
	nReceiverInterrupt int
	nTransmitterPin    int
	timings            [RCSWITCH_MAX_CHANGES]int
}

func NewRCSwitch() (*RCSwitch, error) {
	if err := embd.InitGPIO(); err != nil {
		return nil, fmt.Errorf("init gpio: %s", err)
	}
	// defer embd.CloseGPIO()

	rc := &RCSwitch{}
	rc.nTransmitterPin = -1
	rc.setPulseLength(350)
	rc.setRepeatTransmit(10)
	rc.setProtocol(1)
	rc.nReceiverInterrupt = -1
	rc.setReceiveTolerance(60)
	rc.nReceivedValue = 0
	return rc, nil
}

/**
 * Sets the protocol to send.
 */
func (rc *RCSwitch) setProtocol(nProtocol int) {
	rc.nProtocol = nProtocol
	if nProtocol == 1 {
		rc.setPulseLength(350)
	} else if nProtocol == 2 {
		rc.setPulseLength(650)
	} else if nProtocol == 3 {
		rc.setPulseLength(100)
	} else if nProtocol == 4 {
		rc.setPulseLength(380)
	} else if nProtocol == 5 {
		rc.setPulseLength(500)
	}
}

/**
 * Sets the protocol to send with pulse length in microseconds.
 */
func (rc *RCSwitch) setProtocolAndLength(nProtocol int, nPulseLength int) {
	rc.nProtocol = nProtocol
	rc.setPulseLength(nPulseLength)
}

/**
 * Sets pulse length in microseconds
 */
func (rc *RCSwitch) setPulseLength(nPulseLength int) {
	rc.nPulseLength = nPulseLength
}

/**
 * Sets Repeat Transmits
 */
func (rc *RCSwitch) setRepeatTransmit(nRepeatTransmit int) {
	rc.nRepeatTransmit = nRepeatTransmit
}

/**
 * Set Receiving Tolerance
 */
func (rc *RCSwitch) setReceiveTolerance(nPercent int) {
	rc.nReceiveTolerance = nPercent
}

/**
 * Enable transmissions
 *
 * @param nTransmitterPin    Arduino Pin to which the sender is connected to
 */
func (rc *RCSwitch) enableTransmit(nTransmitterPin int) {
	rc.nTransmitterPin = nTransmitterPin
	embd.SetDirection(rc.nTransmitterPin, embd.Out)
}

/**
 * Disable transmissions
 */
func (rc *RCSwitch) disableTransmit() {
	rc.nTransmitterPin = -1
}

/**
 * Switch a remote switch on (Type D REV)
 *
 * @param sGroup        Code of the switch group (A,B,C,D)
 * @param nDevice       Number of the switch itself (1..3)
 */
func (rc *RCSwitch) switchOnTypeD(sGroup rune, nDevice int) {
	rc.sendTriState(rc.getCodeWordD(sGroup, nDevice, true))
}

/**
 * Switch a remote switch off (Type D REV)
 *
 * @param sGroup        Code of the switch group (A,B,C,D)
 * @param nDevice       Number of the switch itself (1..3)
 */
func (rc *RCSwitch) switchOffTypeD(sGroup rune, nDevice int) {
	rc.sendTriState(rc.getCodeWordD(sGroup, nDevice, false))
}

/**
 * Switch a remote switch on (Type C Intertechno)
 *
 * @param sFamily  Familycode (a..f)
 * @param nGroup   Number of group (1..4)
 * @param nDevice  Number of device (1..4)
 */
func (rc *RCSwitch) switchOnTypeC(sFamily rune, nGroup int, nDevice int) {
	rc.sendTriState(rc.getCodeWordC(sFamily, nGroup, nDevice, true))
}

/**
 * Switch a remote switch off (Type C Intertechno)
 *
 * @param sFamily  Familycode (a..f)
 * @param nGroup   Number of group (1..4)
 * @param nDevice  Number of device (1..4)
 */
func (rc *RCSwitch) switchOffTypeC(sFamily rune, nGroup int, nDevice int) {
	rc.sendTriState(rc.getCodeWordC(sFamily, nGroup, nDevice, false))
}

/**
 * Switch a remote switch on (Type B with two rotary/sliding switches)
 *
 * @param nAddressCode  Number of the switch group (1..4)
 * @param nChannelCode  Number of the switch itself (1..4)
 */
func (rc *RCSwitch) switchOnTypeB(nAddressCode int, nChannelCode int) {
	rc.sendTriState(rc.getCodeWordB(nAddressCode, nChannelCode, true))
}

/**
 * Switch a remote switch off (Type B with two rotary/sliding switches)
 *
 * @param nAddressCode  Number of the switch group (1..4)
 * @param nChannelCode  Number of the switch itself (1..4)
 */
func (rc *RCSwitch) switchOffTypeB(nAddressCode int, nChannelCode int) {
	rc.sendTriState(rc.getCodeWordB(nAddressCode, nChannelCode, false))
}

/**
 * Switch a remote switch on (Type A with 10 pole DIP switches)
 *
 * @param sGroup        Code of the switch group (refers to DIP switches 1..5 where "1" = on and "0" = off, if all DIP switches are on it's "11111")
 * @param sDevice       Code of the switch device (refers to DIP switches 6..10 (A..E) where "1" = on and "0" = off, if all DIP switches are on it's "11111")
 */
func (rc *RCSwitch) switchOnTypeA(sGroup string, sDevice string) {
	rc.sendTriState(rc.getCodeWordA(sGroup, sDevice, true))
}

/**
 * Switch a remote switch off (Type A with 10 pole DIP switches)
 *
 * @param sGroup        Code of the switch group (refers to DIP switches 1..5 where "1" = on and "0" = off, if all DIP switches are on it's "11111")
 * @param sDevice       Code of the switch device (refers to DIP switches 6..10 (A..E) where "1" = on and "0" = off, if all DIP switches are on it's "11111")
 */
func (rc *RCSwitch) switchOffTypeA(sGroup string, sDevice string) {
	rc.sendTriState(rc.getCodeWordA(sGroup, sDevice, false))
}

/**
 * Returns a rune[13], representing the Code Word to be send.
 * A Code Word consists of 9 address bits, 3 data bits and one sync bit but in our case only the first 8 address bits and the last 2 data bits were used.
 * A Code Bit can have 4 different states: "F" (floating), "0" (low), "1" (high), "S" (synchronous bit)
 *
 * +-------------------------------+--------------------------------+-----------------------------------------+-----------------------------------------+----------------------+------------+
 * | 4 bits address (switch group) | 4 bits address (switch number) | 1 bit address (not used, so never mind) | 1 bit address (not used, so never mind) | 2 data bits (on|off) | 1 sync bit |
 * | 1=0FFF 2=F0FF 3=FF0F 4=FFF0   | 1=0FFF 2=F0FF 3=FF0F 4=FFF0    | F                                       | F                                       | on=FF off=F0         | S          |
 * +-------------------------------+--------------------------------+-----------------------------------------+-----------------------------------------+----------------------+------------+
 *
 * @param nAddressCode  Number of the switch group (1..4)
 * @param nChannelCode  Number of the switch itself (1..4)
 * @param bStatus       Wether to switch on (true) or off (false)
 *
 * @return rune[13]
 */
func (rc *RCSwitch) getCodeWordB(nAddressCode int, nChannelCode int, bStatus bool) string {
	word := ""
	code := [5]string{"FFFF", "0FFF", "F0FF", "FF0F", "FFF0"}
	if nAddressCode < 1 || nAddressCode > 4 || nChannelCode < 1 || nChannelCode > 4 {
		return ""
	}
	for i := 0; i < 4; i++ {
		word += string(code[nAddressCode][i])
	}
	for i := 0; i < 4; i++ {
		word += string(code[nChannelCode][i])
	}
	word += "FFF"
	if bStatus {
		word += string('F')
	} else {
		word += string('0')
	}
	return word
}

/**
 * Returns a rune[13], representing the Code Word to be send.
 *
 * getCodeWordA(string, string)
 *
 */
func (rc *RCSwitch) getCodeWordA(sGroup string, sDevice string, bOn bool) string {
	word := ""
	i, j := 0, 0

	for i = 0; i < 5; i++ {
		if sGroup[i] == '0' {
			word += string('F')
		} else {
			word += string('0')
		}
	}

	for i = 0; i < 5; i++ {
		if sDevice[i] == '0' {
			word += string('F')
		} else {
			word += string('0')
		}
	}

	if bOn {
		word += string('0')
		word += string('F')
	} else {
		word += string('F')
		word += string('0')
	}

	return word
}

/**
 * Like getCodeWord (Type C = Intertechno)
 */
func (rc *RCSwitch) getCodeWordC(sFamily rune, nGroup int, nDevice int, bStatus bool) string {
	word := ""
	nReturnPos := 0

	if sFamily < 97 || sFamily > 112 || nGroup < 1 || nGroup > 4 || nDevice < 1 || nDevice > 4 {
		return ""
	}

	sDeviceGroupCode := dec2binWzerofill((nDevice-1)+(nGroup-1)*4, 4)
	familycode := [16]string{"0000", "F000", "0F00", "FF00", "00F0", "F0F0", "0FF0", "FFF0", "000F", "F00F", "0F0F", "FF0F", "00FF", "F0FF", "0FFF", "FFFF"}
	for i := 0; i < 4; i++ {
		word += string(familycode[sFamily-97][i])
	}
	for i := 0; i < 4; i++ {
		if sDeviceGroupCode[3-i] == '1' {
			word += string('F')
		} else {
			word += string('0')
		}
	}
	word += string('0')
	word += string('F')
	word += string('F')
	if bStatus {
		word += string('F')
	} else {
		word += string('0')
	}
	return word
}

/**
 * Decoding for the REV Switch Type
 *
 * Returns a rune[13], representing the Tristate to be send.
 * A Code Word consists of 7 address bits and 5 command data bits.
 * A Code Bit can have 3 different states: "F" (floating), "0" (low), "1" (high)
 *
 * +-------------------------------+--------------------------------+-----------------------+
 * | 4 bits address (switch group) | 3 bits address (device number) | 5 bits (command data) |
 * | A=1FFF B=F1FF C=FF1F D=FFF1   | 1=0FFF 2=F0FF 3=FF0F 4=FFF0    | on=00010 off=00001    |
 * +-------------------------------+--------------------------------+-----------------------+
 *
 * Source: http://www.the-intruder.net/funksteckdosen-von-rev-uber-arduino-ansteuern/
 *
 * @param sGroup        Name of the switch group (A..D, resp. a..d)
 * @param nDevice       Number of the switch itself (1..3)
 * @param bStatus       Whether to switch on (true) or off (false)
 *
 * @return rune[13]
 */

func (rc *RCSwitch) getCodeWordD(sGroup rune, nDevice int, bStatus bool) string {

	nReturnPos := 0

	// Building 4 bits address
	// (Potential problem if dec2binWrunefill not returning correct string)
	sGroupCode := ""
	switch sGroup {
	case 'a':
	case 'A':
		sGroupCode = dec2binWrunefill(8, 4, 'F')
		break
	case 'b':
	case 'B':
		sGroupCode = dec2binWrunefill(4, 4, 'F')
		break
	case 'c':
	case 'C':
		sGroupCode = dec2binWrunefill(2, 4, 'F')
		break
	case 'd':
	case 'D':
		sGroupCode = dec2binWrunefill(1, 4, 'F')
		break
	default:
		return ""
	}

	word := ""
	for i := 0; i < 4; i++ {
		word += string(sGroupCode[i])
	}

	// Building 3 bits address
	// (Potential problem if dec2binWrunefill not returning correct string)
	switch nDevice {
	case 1:
		word += dec2binWrunefill(4, 3, 'F')
		break
	case 2:
		word += dec2binWrunefill(2, 3, 'F')
		break
	case 3:
		word += dec2binWrunefill(1, 3, 'F')
		break
	default:
		return ""
	}

	// fill up rest with zeros
	for i := 0; i < 5; i++ {
		word += string('0')
	}
	// encode on or off
	// if bStatus {
	// 	word[10] = '1'
	// } else {
	// 	word[11] = '1'
	// }
	return word

}

/**
 * @param sCodeWord   /^[10FS]*$/  -> see getCodeWord
 */
func (rc *RCSwitch) sendTriState(sCodeWord string) {
	for nRepeat := 0; nRepeat < rc.nRepeatTransmit; nRepeat++ {
		for r := range sCodeWord {
			switch r {
			case '0':
				rc.sendT0()
				break
			case 'F':
				rc.sendTF()
				break
			case '1':
				rc.sendT1()
				break
			}
		}
		rc.sendSync()
	}
}

func (rc *RCSwitch) sendCode(Code int, length int) {
	rc.send(rc.dec2binWzerofill(Code, length))
}

func (rc *RCSwitch) send(sCodeWord string) {
	for nRepeat := 0; nRepeat < rc.nRepeatTransmit; nRepeat++ {
		for r := range sCodeWord {
			switch r {
			case '0':
				rc.send0()
				break
			case '1':
				rc.send1()
				break
			}
		}
		rc.sendSync()
	}
}

func (rc *RCSwitch) transmit(nHighPulses int, nLowPulses int) {

	disabledReceive := false
	prevReceiverInterrupt := rc.nReceiverInterrupt

	if rc.nTransmitterPin != -1 {

		if rc.nReceiverInterrupt != -1 {
			rc.disableReceive()
			disabledReceive = true
		}

		embd.DigitalWrite(rc.nTransmitterPin, embd.High)
		time.Sleep(time.Duration(rc.nPulseLength*nHighPulses) * time.Microsecond)
		embd.DigitalWrite(rc.nTransmitterPin, embd.Low)
		time.Sleep(time.Duration(rc.nPulseLength*nLowPulses) * time.Microsecond)

		if disabledReceive {
			rc.enableReceiveInterrupt(prevReceiverInterrupt)
		}

	}
}

/**
 * Sends a "0" Bit
 *                       _
 * Waveform Protocol 1: | |___
 *                       _
 * Waveform Protocol 2: | |__
 */
func (rc *RCSwitch) send0() {
	if rc.nProtocol == 1 {
		rc.transmit(1, 3)
	} else if rc.nProtocol == 2 {
		rc.transmit(1, 2)
	} else if rc.nProtocol == 3 {
		rc.transmit(4, 11)
	} else if rc.nProtocol == 4 {
		rc.transmit(1, 3)
	} else if rc.nProtocol == 5 {
		rc.transmit(1, 2)
	}
}

/**
 * Sends a "1" Bit
 *                       ___
 * Waveform Protocol 1: |   |_
 *                       __
 * Waveform Protocol 2: |  |_
 */
func (rc *RCSwitch) send1() {
	if rc.nProtocol == 1 {
		rc.transmit(3, 1)
	} else if rc.nProtocol == 2 {
		rc.transmit(2, 1)
	} else if rc.nProtocol == 3 {
		rc.transmit(9, 6)
	} else if rc.nProtocol == 4 {
		rc.transmit(3, 1)
	} else if rc.nProtocol == 5 {
		rc.transmit(2, 1)
	}
}

/**
 * Sends a Tri-State "0" Bit
 *            _     _
 * Waveform: | |___| |___
 */
func (rc *RCSwitch) sendT0() {
	rc.transmit(1, 3)
	rc.transmit(1, 3)
}

/**
 * Sends a Tri-State "1" Bit
 *            ___   ___
 * Waveform: |   |_|   |_
 */
func (rc *RCSwitch) sendT1() {
	rc.transmit(3, 1)
	rc.transmit(3, 1)
}

/**
 * Sends a Tri-State "F" Bit
 *            _     ___
 * Waveform: | |___|   |_
 */
func (rc *RCSwitch) sendTF() {
	rc.transmit(1, 3)
	rc.transmit(3, 1)
}

/**
 * Sends a "Sync" Bit
 *                       _
 * Waveform Protocol 1: | |_______________________________
 *                       _
 * Waveform Protocol 2: | |__________
 */
func (rc *RCSwitch) sendSync() {

	if rc.nProtocol == 1 {
		rc.transmit(1, 31)
	} else if rc.nProtocol == 2 {
		rc.transmit(1, 10)
	} else if rc.nProtocol == 3 {
		rc.transmit(1, 71)
	} else if rc.nProtocol == 4 {
		rc.transmit(1, 6)
	} else if rc.nProtocol == 5 {
		rc.transmit(6, 14)
	}
}

/**
 * Enable receiving data
 */
func (rc *RCSwitch) enableReceiveInterrupt(interrupt int) {
	rc.nReceiverInterrupt = interrupt
	rc.enableReceive()
}

func (rc *RCSwitch) enableReceive() {
	if rc.nReceiverInterrupt != -1 {
		rc.nReceivedValue = 0
		rc.nReceivedBitlength = 0
		rc.attachInterrupt(rc.nReceiverInterrupt, handleInterrupt, CHANGE)

	}
}

/**
 * Disable receiving data
 */
func (rc *RCSwitch) disableReceive() {
	detachInterrupt(rc.nReceiverInterrupt)
	rc.nReceiverInterrupt = -1
}

func (rc *RCSwitch) available() bool {
	return rc.nReceivedValue != 0
}

func (rc *RCSwitch) resetAvailable() {
	rc.nReceivedValue = 0
}

func (rc *RCSwitch) getReceivedValue() int {
	return rc.nReceivedValue
}

func (rc *RCSwitch) getReceivedBitlength() int {
	return rc.nReceivedBitlength
}

func (rc *RCSwitch) getReceivedDelay() int {
	return rc.nReceivedDelay
}

func (rc *RCSwitch) getReceivedProtocol() int {
	return rc.nReceivedProtocol
}

func (rc *RCSwitch) getReceivedRawdata() []int {
	return rc.timings
}

/**
 *
 */
func (rc *RCSwitch) receiveProtocol1(changeCount int) bool {

	code := 0
	delay := rc.timings[0] / 31
	delayTolerance := delay * rc.nReceiveTolerance * 0.01

	for i := 1; i < changeCount; i = i + 2 {

		if rc.timings[i] > delay-delayTolerance && rc.timings[i] < delay+delayTolerance && rc.timings[i+1] > delay*3-delayTolerance && rc.timings[i+1] < delay*3+delayTolerance {
			code = code << 1
		} else if rc.timings[i] > delay*3-delayTolerance && rc.timings[i] < delay*3+delayTolerance && rc.timings[i+1] > delay-delayTolerance && rc.timings[i+1] < delay+delayTolerance {
			code += 1
			code = code << 1
		} else {
			// Failed
			i = changeCount
			code = 0
		}
	}
	code = code >> 1
	if changeCount > 6 { // ignore < 4bit values as there are no devices sending 4bit values => noise
		rc.nReceivedValue = code
		rc.nReceivedBitlength = changeCount / 2
		rc.nReceivedDelay = delay
		rc.nReceivedProtocol = 1
	}

	if code == 0 {
		return false
	} else if code != 0 {
		return true
	}

}

func (rc *RCSwitch) receiveProtocol2(changeCount int) bool {

	code := 0
	delay := rc.timings[0] / 10
	delayTolerance := delay * rc.nReceiveTolerance * 0.01

	for i := 1; i < changeCount; i = i + 2 {

		if rc.timings[i] > delay-delayTolerance && rc.timings[i] < delay+delayTolerance && rc.timings[i+1] > delay*2-delayTolerance && rc.timings[i+1] < delay*2+delayTolerance {
			code = code << 1
		} else if rc.timings[i] > delay*2-delayTolerance && rc.timings[i] < delay*2+delayTolerance && rc.timings[i+1] > delay-delayTolerance && rc.timings[i+1] < delay+delayTolerance {
			code += 1
			code = code << 1
		} else {
			// Failed
			i = changeCount
			code = 0
		}
	}
	code = code >> 1
	if changeCount > 6 { // ignore < 4bit values as there are no devices sending 4bit values => noise
		rc.nReceivedValue = code
		rc.nReceivedBitlength = changeCount / 2
		rc.nReceivedDelay = delay
		rc.nReceivedProtocol = 2
	}

	if code == 0 {
		return false
	} else if code != 0 {
		return true
	}

}

/** Protocol 3 is used by BL35P02.
 *
 */
func (rc *RCSwitch) receiveProtocol3(changeCount int) bool {

	code := 0
	delay := rc.timings[0] / PROTOCOL3_SYNC_FACTOR
	delayTolerance := delay * rc.nReceiveTolerance * 0.01

	for i := 1; i < changeCount; i = i + 2 {

		if rc.timings[i] > delay*PROTOCOL3_0_HIGH_CYCLES-delayTolerance &&
			rc.timings[i] < delay*PROTOCOL3_0_HIGH_CYCLES+delayTolerance &&
			rc.timings[i+1] > delay*PROTOCOL3_0_LOW_CYCLES-delayTolerance &&
			rc.timings[i+1] < delay*PROTOCOL3_0_LOW_CYCLES+delayTolerance {
			code = code << 1
		} else if rc.timings[i] > delay*PROTOCOL3_1_HIGH_CYCLES-delayTolerance &&
			rc.timings[i] < delay*PROTOCOL3_1_HIGH_CYCLES+delayTolerance &&
			rc.timings[i+1] > delay*PROTOCOL3_1_LOW_CYCLES-delayTolerance &&
			rc.timings[i+1] < delay*PROTOCOL3_1_LOW_CYCLES+delayTolerance {
			code += 1
			code = code << 1
		} else {
			// Failed
			i = changeCount
			code = 0
		}
	}
	code = code >> 1
	if changeCount > 6 { // ignore < 4bit values as there are no devices sending 4bit values => noise
		rc.nReceivedValue = code
		rc.nReceivedBitlength = changeCount / 2
		rc.nReceivedDelay = delay
		rc.nReceivedProtocol = 3
	}

	if code == 0 {
		return false
	} else if code != 0 {
		return true
	}
}

func (rc *RCSwitch) handleInterrupt() {

	time := micros()
	var duration int = time - lastTime
	var changeCount int
	var lastTime int
	var repeatCount int

	if duration > rc.nSeparationLimit && duration > rc.timings[0]-200 && duration < rc.timings[0]+200 {
		repeatCount++
		changeCount--
		if repeatCount == 2 {
			if receiveProtocol1(changeCount) == false {
				if receiveProtocol2(changeCount) == false {
					if receiveProtocol3(changeCount) == false {
						//failed
					}
				}
			}
			repeatCount = 0
		}
		changeCount = 0
	} else if duration > rc.nSeparationLimit {
		changeCount = 0
	}

	if changeCount >= RCSWITCH_MAX_CHANGES {
		changeCount = 0
		repeatCount = 0
	}
	rc.timings[changeCount] = duration
	changeCount++
	lastTime = time
}

/**
 * Turns a decimal value to its binary representation
 */
func (rc *RCSwitch) dec2binWzerofill(dec int, bitLength int) string {
	return dec2binWrunefill(dec, bitLength, '0')
}

func (rc *RCSwitch) dec2binWrunefill(dec int, bitLength int, fill rune) string {
	bin := [64]rune{}
	i := 0

	for dec > 0 {

		if (dec & 1) > 0 {
			bin[32+i] = '1'

		} else {
			bin[32+i] = fill
		}
		i++
		dec = dec >> 1
	}

	for j := 0; j < bitLength; j++ {
		if j >= bitLength-i {
			bin[j] = bin[31+i-(j-(bitLength-i))]
		} else {
			bin[j] = fill
		}
	}
	bin[bitLength] = ""
	return string(bin)
}
