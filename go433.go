package go433

import (
	"errors"
	"log"
	"time"
)

//Debug controls whether debug prints are emitted
var Debug = false

//Send a 433 code (32bits) out the specified GPIO pin
func Send(pin int, code uint32) error {
	return SendWith(SendOpts{pin, code, 24, 350, 8})
}

//SendOpts for use with SendWith
type SendOpts struct {
	Pin             int
	Code            uint32
	Width           int
	PulseLength     int
	Retransmissions int
}

//SendWith a 433 code (32bits) out the specified GPIO pin with custom options
func SendWith(opts SendOpts) error {
	p, err := openPinOut(opts.Pin)
	if err != nil {
		return err
	}
	//always 1 for now
	const protocol = 1
	//encode hi/los
	transmit := func(hi, lo int) {
		//on for hi-many pulses
		p.Write(true)
		time.Sleep(time.Duration(opts.PulseLength*hi) * time.Microsecond)
		//off for lo-many pulses
		p.Write(false)
		time.Sleep(time.Duration(opts.PulseLength*lo) * time.Microsecond)
	}
	//send 1 bit
	sendBit := func(on bool) {
		switch protocol {
		case 1:
			//protocol 1
			if on {
				transmit(3, 1)
			} else {
				transmit(1, 3)
			}
		default:
			log.Fatal("not supported")
		}
	}
	//get width
	width := uint32(opts.Width)
	if width < 4 {
		width = 24
		//auto-calculate width (min 4 bits)
		//NOTE disabled
		// for width = uint32(4); (uint32(1) << width) < opts.Code; width++ {}
	}
	if opts.Code > uint32(uint64(1<<width)-1) {
		return errors.New("code cannot fit within given width")
	}
	if Debug {
		log.Printf("transmission width: %d", width)
	}
	//retranmissions
	for r := 0; r < opts.Retransmissions; r++ {
		//send all bits
		mask := uint32(1) << (width - 1)
		for shift := uint32(0); shift < width; shift++ {
			set := (opts.Code & mask) > 0
			sendBit(set)
			mask = mask >> 1
		}
		//send sync
		if protocol == 1 {
			transmit(1, 31)
		}
	}
	//done!
	return nil
}

//Receive codes from the given pin, close the channel
//to stop receiving.
func Receive(pin int, handler func(code uint32)) (cancel func(), err error) {
	p, err := openPinIn(pin)
	if err != nil {
		return nil, err
	}
	ch, err := p.OnChange()
	if err != nil {
		return nil, err
	}
	//receive state
	const MaxChanges = 67
	const ReceiveTolerance = 60
	const separationLimit = 5 * time.Millisecond
	const separationError = 200 * time.Microsecond
	changeCount := 0
	repeatCount := 0
	timings := [MaxChanges]time.Duration{}
	//a is within b, given error bound
	within := func(a, b, e time.Duration) bool {
		return a > b-e && a < b+e
	}
	//decode captured timings
	decode := func() bool {
		code := uint32(0)
		delay := timings[0] / 31
		delay3 := delay * 3
		delayTolerance := delay * time.Duration(ReceiveTolerance) / 100
		if Debug {
			log.Printf("decode - changes: %d, delay: %s, tolerance: %s", changeCount, delay, delayTolerance)
		}
		for i := 1; i < changeCount; i = i + 2 {
			this := timings[i]
			next := timings[i+1]
			//increasing delay
			inc := within(this, delay, delayTolerance) && within(next, delay3, delayTolerance)
			//decreasing delay
			dec := within(this, delay3, delayTolerance) && within(next, delay, delayTolerance)
			log.Printf(">>>> decode - inc: %v, dec: %v, code: %d", inc, dec, code)
			if inc {
				code = code << 1
			} else if dec {
				code++
				code = code << 1
			} else {
				return false
			}
		}
		code = code >> 1
		bits := changeCount / 2
		if bits >= 4 {
			if Debug {
				log.Printf("success - delay: %s, bits: %d, code: %d", delay, bits, code)
			}
			handler(code)
		}
		return code != 0
	}
	//pin interupt handler
	last := time.Now()
	handle := func(on bool) {
		now := time.Now()
		delta := now.Sub(last)
		last = now
		if delta > separationLimit {
			lower := timings[0] - separationError
			upper := timings[0] + separationError
			w := delta > lower && delta < upper
			if Debug {
				log.Printf("handle: %s < %s < %s = %v", lower, delta, upper, w)
			}
			if w {
				repeatCount++
				changeCount--
				if repeatCount == 2 {
					decode()
					repeatCount = 0
				}
			}
			changeCount = 0
		}
		if changeCount >= MaxChanges {
			changeCount = 0
			repeatCount = 0
		}
		timings[changeCount] = delta
		changeCount++
	}

	go func() {
		for on := range ch {
			handle(on)
		}
	}()
	return func() {
		close(ch)
	}, nil
}
