package main

import (
	"log"
	"time"

	"github.com/jpillora/opts"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
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
		Repeat:      3,
		OnDuration:  250 * time.Millisecond,
		OffDuration: 250 * time.Millisecond,
	}
	opts.Parse(&c)

	embd.InitGPIO()
	defer embd.CloseGPIO()

	pin, err := embd.NewDigitalPin(c.Pin)
	if err != nil {
		log.Fatal(err)
	}
	defer pin.Close()

	log.Printf("using pin: %d", c.Pin)
	pin.SetDirection(embd.Out)

	for r := 0; r < c.Repeat; r++ {
		pin.Write(embd.High)
		time.Sleep(c.OnDuration)
		pin.Write(embd.Low)
		time.Sleep(c.OffDuration)
	}

	log.Printf("done")
}
