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
	}
)


func (ic *IC) listenMcp23008Interrupt(iot *Iot, interrupt chan byte) {
	for {
		gpio := <-interrupt
		state := ic.readGPIO(gpio)

		endpoint := iot.conf.Address
		endpoint.Type = kite.H_ENDPOINT
		endpoint.Address = fmt.Sprintf("%d",ic.Address)
		endpoint.Id = fmt.Sprintf("%d", gpio)

		//var message = kite.Message{Data: fmt.Sprintf("Interrupt occurs on %d, new value %d", gpio, state), Sender: iot.conf.Address, Receiver: kite.Address{Domain: iot.conf.Address.Domain, Type: kite.H_ANY, Host: "*", Address: "*", Id: "*"}, Action: kite.A_NOTIFY}
		var message = kite.Message{Data: fmt.Sprintf("Interrupt occurs on %d, new value %d", gpio, state), Sender: endpoint, Receiver: kite.Address{Domain: iot.conf.Address.Domain, Type: kite.H_ANY, Host: "*", Address: "*", Id: "*"}, Action: kite.A_NOTIFY}
		iot.conn.WriteJSON(message)

		message.Action = kite.A_LOG
		iot.conn.WriteJSON(message)
	}
}
func (ic *IC) readGPIO(gpio byte) int {
	return int(mcp23008.ReadGpio(ic.IC.(*mcp23008.Mcp23008), gpio))
}