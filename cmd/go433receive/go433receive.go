package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

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
	t0 := time.Now()
	handle := func() {
		t1 := time.Now()
		delta := t1.Sub(t0)
		t0 = t1
		fmt.Printf("[%05d] %s\n", i, delta)
		i++
	}

	if err := pin.BeginWatch(gpio.EdgeBoth, handle); err != nil {
		fmt.Printf("Unable to watch pin: %s\n", err.Error())
		os.Exit(1)
	}

	log.Printf("watching %d", c.Pin)
	select {}
}

// "github.com/hugozhu/rpi"
//
// func main() {
// 	c := config{
// 		Pin: rpi.PIN_GPIO_2,
// 	}
// 	opts.Parse(&c)
// 	if err := rpi.WiringPiSetup(); err != nil {
// 		log.Fatal(err)
// 	}
// 	interupts := rpi.WiringPiISR(c.Pin, rpi.INT_EDGE_BOTH)
// 	log.Printf("watching %d", c.Pin)
// 	i := 0
// 	t0 := time.Now()
// 	for range interupts {
// 		t1 := time.Now()
// 		delta := t1.Sub(t0)
// 		t0 = t1
// 		fmt.Printf("[%05d] %s\n", i, delta)
// 		i++
// 	}
// }
