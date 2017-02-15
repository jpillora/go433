package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/jpillora/go433"
	"github.com/jpillora/opts"
)

type config struct {
	Pin   int
	Debug bool
}

func main() {
	c := config{Pin: 27}
	opts.Parse(&c)

	go433.Debug = c.Debug

	receiving, err := go433.Receive(c.Pin, func(code uint32) {
		log.Printf("pin: %d, received: %d", c.Pin, code)
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("pin: %d, waiting to receive...", c.Pin)

	//Ctrl+C to stop
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	close(receiving)
}
