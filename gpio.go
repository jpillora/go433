package go433

//PinOut provides GPIO out
type PinOut interface {
	Write(on bool) error
}

//PinIn provides GPIO in
type PinIn interface {
	OnChange() (chan bool, error)
}
