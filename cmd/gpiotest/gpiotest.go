package main

import (
	"log"
	"time"

	"github.com/jpillora/opts"
	"github.com/stianeikeland/go-rpio"
)

type config struct {
	Pin         int
	Repeat      int
	OnDuration  time.Duration
	OffDuration time.Duration
}

func main() {

	c := config{
		Pin:         20,
		Repeat:      1,
		OnDuration:  750 * time.Millisecond,
		OffDuration: 250 * time.Millisecond,
	}
	opts.Parse(&c)

	if err := rpio.Open(); err != nil {
		log.Fatal(err)
	}
	log.Printf("using pin: %d", c.Pin)
	pin := rpio.Pin(c.Pin)
	pin.Output()

	for r := 0; r < c.Repeat; r++ {
		pin.High()
		time.Sleep(c.OnDuration)
		pin.Low()
		time.Sleep(c.OffDuration)
	}

	log.Printf("done")
}
