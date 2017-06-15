// +build linux,arm

package go433

import (
	"errors"
	"strconv"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

var hostInit = errors.New("not inited")

func init() {
	_, err := host.Init()
	hostInit = err
}

//==============

type rpiPinIn struct {
	periph gpio.PinIn
}

func OpenPinIn(n int) (PinIn, error) {
	if hostInit != nil {
		return nil, hostInit
	}
	p := gpioreg.ByName(strconv.Itoa(n))
	if p == nil {
		return nil, errors.New("invalid pin")
	}
	if err := p.In(gpio.Float, gpio.BothEdges); err != nil {
		return nil, err
	}
	return &rpiPinIn{
		periph: gpio.PinIn(p),
	}, nil
}

func (r *rpiPinIn) OnChange() (chan bool, error) {
	changes := make(chan bool)
	noTimeout := time.Duration(-1)
	go func() {
		for {
			on := r.periph.WaitForEdge(noTimeout)
			changes <- on
		}
	}()
	return changes, nil
}

//==============

type rpiPinOut struct {
	periph gpio.PinOut
}

func OpenPinOut(n int) (PinOut, error) {
	if hostInit != nil {
		return nil, hostInit
	}
	p := gpioreg.ByName(strconv.Itoa(n))
	if p == nil {
		return nil, errors.New("invalid pin")
	}
	if err := p.Out(gpio.Low); err != nil {
		return nil, err
	}
	return &rpiPinOut{
		periph: gpio.PinOut(p),
	}, nil
}

func (r *rpiPinOut) Write(on bool) error {
	return r.periph.Out(gpio.Level(on))
}
