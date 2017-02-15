package go433

import (
	"errors"
	"log"
	"time"

	"github.com/davecheney/gpio"
	"github.com/davecheney/gpio/rpi"
	rpio "github.com/stianeikeland/go-rpio"
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
	if err := rpio.Open(); err != nil {
		return err
	}
	p := rpio.Pin(opts.Pin)
	p.Output()
	on := func(on bool) {
		if on {
			p.High()
		} else {
			p.Low()
		}
	}

	//TODO: why doesn't davecheney's package work!?!?!
	// p, err := rpi.OpenPin(pin, gpio.ModeOutput)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// on := func(on bool) {
	// 	for i := 0; i < 100; i++ {
	// 		if on {
	// 			p.Set()
	// 		} else {
	// 			p.Clear()
	// 		}
	// 	}
	// }

	//encode hi/los
	transmit := func(hi, lo int) {
		//on for hi-many pulses
		on(true)
		time.Sleep(time.Duration(opts.PulseLength*hi) * time.Microsecond)
		//off for lo-many pulses
		on(false)
		time.Sleep(time.Duration(opts.PulseLength*lo) * time.Microsecond)
	}

	//always 1 for now
	const protocol = 1

	//send 1 bit
	sendBit := func(on bool) {
		switch protocol {
		case 1:
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
	//send all bits
	for r := 0; r < opts.Retransmissions; r++ {
		//send bits
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
func Receive(pin int, handler func(code uint32)) (chan bool, error) {
	p, err := rpi.OpenPin(pin, gpio.ModeInput)
	if err != nil {
		return nil, err
	}
	//receive state
	const MaxChanges = 67
	const ReceiveTolerance = 60
	changeCount := 0
	repeatCount := 0
	timings := [MaxChanges]time.Duration{}
	separationLimit := 5 * time.Millisecond
	separationError := 200 * time.Microsecond
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
			//increasing delay
			inc := within(timings[i], delay, delayTolerance) && within(timings[i+1], delay3, delayTolerance)
			//decreasing delay
			dec := within(timings[i], delay3, delayTolerance) && within(timings[i+1], delay, delayTolerance)
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
	handle := func() {
		now := time.Now()
		delta := now.Sub(last)
		last = now
		if delta > separationLimit {
			// micros := delta.Nanoseconds() / int64(1000)
			// if micros > 1000 {
			// 	fmt.Printf("[%05d] %d\n", i, micros)
			// 	i++
			// }
			if delta > timings[0]-separationError && delta < timings[0]+separationError {
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
	if err := p.BeginWatch(gpio.EdgeBoth, handle); err != nil {
		p.Close()
		return nil, err
	}
	ch := make(chan bool)
	go func() {
		<-ch
		p.Close()
		p.EndWatch()
	}()
	return ch, nil
}
