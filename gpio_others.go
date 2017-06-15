// +build !linux !arm

package go433

import "errors"

func openPinIn(n int) (pinIn, error) {
	return nil, errors.New("not available on this os/platform")
}

func openPinOut(n int) (pinOut, error) {
	return nil, errors.New("not available on this os/platform")
}
