// +build !linux !arm

package go433

import "errors"

//OpenPinIn initialises a GPIO input pin
func OpenPinIn(n int) (PinIn, error) {
	return nil, errors.New("not available on this os/platform")
}

//OpenPinOut initialises a GPIO output pin
func OpenPinOut(n int) (PinOut, error) {
	return nil, errors.New("not available on this os/platform")
}
