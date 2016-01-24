package main

import (
	"log"
	"strconv"
	"time"

	"github.com/davecheney/gpio"
	"github.com/davecheney/gpio/rpi"
	"github.com/jpillora/opts"
)

type config struct {
	Code        string `type:"arg"`
	Pin         int
	PulseLength int
	Retransmit  int
	Protocol    int
}

func main() {

	c := config{
		Pin:         17,
		PulseLength: 350,
		Retransmit:  10,
		Protocol:    1,
	}
	opts.Parse(&c)

	code, err := strconv.ParseUint(c.Code, 10, 32)
	if err != nil {
		log.Fatal(err)
	}

	//GPIO SETUP
	pin, err := rpi.OpenPin(c.Pin, gpio.ModeOutput)
	if err != nil {
		log.Fatal(err)
	}

	transmit := func(hi, lo int) {
		// fmt.Printf("H%d", hi)
		pin.Set()
		time.Sleep(time.Duration(c.PulseLength*hi) * time.Microsecond)
		// fmt.Printf("L%d\n", lo)
		pin.Clear()
		time.Sleep(time.Duration(c.PulseLength*lo) * time.Microsecond)
	}

	sendBit := func(on bool) {
		switch c.Protocol {
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

	send := func(code, width uint32) {
		for r := 0; r < c.Retransmit; r++ {
			mask := uint32(1) << (width - 1)
			for shift := uint32(0); shift < width; shift++ {
				set := (code & mask) > 0
				sendBit(set)
				// log.Printf("shift %d [%d & %d] = %v", shift, code, mask, set)
				mask = mask >> 1
			}
			//send sync
			if c.Protocol == 1 {
				transmit(1, 31)
			}
			// log.Printf("transmitted (#%d)", r+1)
		}
	}

	t0 := time.Now()
	log.Printf("sending on pin %d (pulse %s)", c.Pin, time.Duration(c.PulseLength)*time.Microsecond)
	send(uint32(code), 24)
	log.Printf("sent: %d (in %s)", code, time.Now().Sub(t0))
}
