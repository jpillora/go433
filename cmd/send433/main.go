package main

import (
	"log"
	"strconv"

	"github.com/jpillora/go433"
	"github.com/jpillora/opts"
)

type config struct {
	Code            string `type:"arg"`
	Pin             int
	Width           int
	PulseLength     int
	Retransmissions int
	Debug           bool
}

func main() {
	c := config{}
	c.Pin = 17
	c.PulseLength = 350
	c.Retransmissions = 8
	opts.Parse(&c)

	go433.Debug = c.Debug

	code, err := strconv.ParseUint(c.Code, 10, 32)
	if err != nil {
		log.Fatal(err)
	}
	if err := go433.SendWith(go433.SendOpts{
		c.Pin, uint32(code), c.Width, c.PulseLength, c.Retransmissions,
	}); err != nil {
		log.Fatal(err)
	}
	log.Printf("pin: %d, sent: %d", c.Pin, code)
}
