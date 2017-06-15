package go433

type pinOut interface {
	Write(on bool) error
}

type pinIn interface {
	OnChange() (chan bool, error)
}
