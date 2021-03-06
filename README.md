# go433

Send and receive 433 MHz using a Raspberry Pi and Go

### Features

* Simple
* No dependencies (WiringPi is not required)

### Usage

Send

```go
//send 1234 on GPIO pin 17
if err := go433.Send(17, 1234); err != nil {
	log.Fatal(err)
}
```

Receive

```go
//receive on GPIO pin 27
receiving, err := go433.Receive(27, func(code uint32) {
	//code is 1234!
})
if err != nil {
	log.Fatal(err)
}
...
close(receiving)
```

*See pinout diagram below*

### Notes

* Tested using RaspberryPi 2, using the transmitter/receiver found in [this article](http://www.princetronics.com/how-to-read-433-mhz-codes-w-raspberry-pi-433-mhz-receiver/)
* Code ported from https://github.com/sui77/rc-switch

### RPI2 Pinout

![rp2_pinout](https://cloud.githubusercontent.com/assets/633843/22986824/d7df1830-f400-11e6-81cd-78a7ddf7a080.png)

#### MIT License

Copyright © 2015 Jaime Pillora &lt;dev@jpillora.com&gt;

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
'Software'), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
