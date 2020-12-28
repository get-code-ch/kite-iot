package main

import (
	"fmt"
	kite "github.com/get-code-ch/kite-common"
	"github.com/get-code-ch/mcp23008/v3"
)

type (
	IC struct {
		Address int
		Type kite.IcRef
		IC interface{}
		//interrupt chan byte
	}
)


func (ic *IC) listenMcp23008Interrupt(iot *Iot, interrupt chan byte) {
	for {
		gpio := <-interrupt
		state := ic.readGPIO(gpio)
		//log.Printf("Interrupt occurs on %d, new value %d", gpio, state)

		var message = kite.Message{Data: fmt.Sprintf("Interrupt occurs on %d, new value %d", gpio, state), Sender: iot.conf.Address, Receiver: kite.Address{Domain: iot.conf.Address.Domain, Type: kite.H_ANY, Host: "*", Address: "*", Id: "*"}, Action: kite.A_NOTIFY}
		iot.conn.WriteJSON(message)

		message.Action = kite.A_LOG
		iot.conn.WriteJSON(message)
	}
}
func (ic *IC) readGPIO(gpio byte) int {
	return int(mcp23008.ReadGpio(ic.IC.(*mcp23008.Mcp23008), gpio))
}