package main

import (
	"fmt"
	kite "github.com/get-code-ch/kite-common"
	"github.com/get-code-ch/mcp23008/v3"
	"log"
	"sync"
)

type (
	IC struct {
		address int
		icRef   kite.IcRef
		ic      interface{}
		wg      sync.WaitGroup
	}
)


func (ic *IC) listenMcp23008Interrupt(iot *Iot, interrupt chan byte) {
	log.Printf("Listening M23008 Interrupt started (%d)...", ic.address)
	for {
		gpio := <-interrupt
		state := ic.readGPIO(gpio)

		endpoint := iot.conf.Address
		endpoint.Type = kite.H_ENDPOINT
		endpoint.Address = fmt.Sprintf("%d",ic.address)
		endpoint.Id = fmt.Sprintf("%d", gpio)


		//var message = kite.Message{Data: fmt.Sprintf("Interrupt occurs on %d, new value %d", gpio, state), Sender: iot.conf.address, Receiver: kite.address{Domain: iot.conf.address.Domain, icRef: kite.H_ANY, Host: "*", address: "*", Id: "*"}, Action: kite.A_NOTIFY}
		iot.sync.Lock()
		var message = kite.Message{Data: fmt.Sprintf("Interrupt occurs on %d, new value %d", gpio, state), Sender: endpoint, Receiver: kite.Address{Domain: iot.conf.Address.Domain, Type: kite.H_ANY, Host: "*", Address: "*", Id: "*"}, Action: kite.A_NOTIFY}
		if err := iot.conn.WriteJSON(message); err != nil {
			iot.conn.Close()
			break
		}

		message.Action = kite.A_LOG
		if err := iot.conn.WriteJSON(message); err != nil {
			iot.conn.Close()
			break
		}
		iot.sync.Unlock()
	}
	log.Printf("Listening M23008 Interrupt exited (%d)...", ic.address)
}
func (ic *IC) readGPIO(gpio byte) int {
	return int(mcp23008.ReadGpio(ic.ic.(*mcp23008.Mcp23008), gpio))
}

func (ic *IC) writeGPIO(gpio int, state int) int {
	if state == 0 {
		mcp23008.GpioOff(ic.ic.(*mcp23008.Mcp23008), byte(gpio))
	} else {
		mcp23008.GpioOn(ic.ic.(*mcp23008.Mcp23008), byte(gpio))
	}
	return int(mcp23008.ReadGpio(ic.ic.(*mcp23008.Mcp23008), byte(gpio)))
}
