package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/davecheney/gpio"
	"github.com/davecheney/gpio/rpi"
	"github.com/jpillora/opts"
)

type config struct {
	Pin int
}

func main() {

	c := config{
		Pin: 27,
	}

	opts.Parse(&c)

	pin, err := rpi.OpenPin(c.Pin, gpio.ModeInput)
	if err != nil {
		log.Fatal(err)
	}

	// clean up on exit
	exit := func() {
		fmt.Println("Closing pin")
		pin.Close()
		os.Exit(0)
	}
	defer exit()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		exit()
	}()

	i := 0
	handle := func() {
		fmt.Printf("[%05d] Pin %d = %v\n", i, c.Pin, pin.Get())
		i++
	}

	if err := pin.BeginWatch(gpio.EdgeBoth, handle); err != nil {
		fmt.Printf("Unable to watch pin: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("Now watching pin %d\n", c.Pin)
	select {}
}
