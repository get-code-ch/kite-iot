package main

import (
	"fmt"
	kite "github.com/get-code-ch/kite-common"
	"github.com/get-code-ch/mcp23008/v3"
	"log"
	"strconv"
)

func (ic *IC) listenMcp23008Interrupt(iot *Iot, interrupt chan byte) {
	log.Printf("Listening M23008 Interrupt started (%d)...", ic.address)
	for {
		gpio := <-interrupt
		state := ic.readGPIO(int(gpio))

		address := iot.conf.Address
		address.Type = kite.H_ENDPOINT
		address.Address = fmt.Sprintf("%d", ic.address)
		address.Id = fmt.Sprintf("%d", gpio)

		endpoint := iot.endpoints[address].endpoint
		mode := ""
		slave := ""

		switch endpoint.Attributes["mode"].(type) {
		case string:
			mode = endpoint.Attributes["mode"].(string)
		}
		switch endpoint.Attributes["slave"].(type) {
		case string:
			slave = endpoint.Attributes["slave"].(string)
		}

		iot.sync.Lock()
		switch mode {
		case "event":
			action := kite.Action(endpoint.Attributes["action"].(string))
			command := endpoint.Attributes["data"].(string)

			//log.Printf("event occurs on %v", endpoint)
			//log.Printf("command %v", command)
			//log.Printf("state %v", state)

			if state == 0 && action == kite.A_CMD && command == "reverse" {
				break
			}

			if command == "{state?on:off}" {
				if state == 1 {
					command = "on"
				} else {
					command = "off"
				}
			}

			to := kite.Address{}
			to.StringToAddress(endpoint.Attributes["to"].(string))
			response := kite.Message{Data: command, Sender: address, Receiver: to, Action: action}
			if err := iot.conn.WriteJSON(response); err != nil {
				iot.conn.Close()
			}
			break
		default:
			var response = kite.Message{Sender: endpoint.Address}
			response.Receiver = kite.Address{Domain: iot.conf.Address.Domain, Type: kite.H_ANY, Host: "*", Address: "*", Id: "*"}
			response.Action = kite.A_VALUE

			data := make(map[string]interface{})
			data["type"] = "gpio"
			data["value"] = state == 1
			data["name"] = endpoint.Name
			data["description"] = endpoint.Description

			response.Data = data

			if err := iot.conn.WriteJSON(response); err != nil {
				iot.conn.Close()
				break
			}

			if endpoint.Notification.Telegram {
				response.Data = fmt.Sprintf("New value for GPIO %s (%s) -> %t", endpoint.Description, endpoint.Address, state == 1)
				response.Action = kite.A_NOTIFY
				response.Receiver = kite.Address{Domain: "telegram", Type: kite.H_ANY, Host: "*", Address: "*", Id: "*"}
				if err := iot.conn.WriteJSON(response); err != nil {
					iot.conn.Close()
					break
				}
			}

			// if GPIO had a slave we copy state to it and we send notification
			if slave != "" {
				if gpio, err := strconv.Atoi(slave); err == nil {
					ic.writeGPIO(gpio, state)
				}
				response.Receiver = kite.Address{Domain: iot.conf.Address.Domain, Type: kite.H_ANY, Host: "*", Address: "*", Id: "*"}
				response.Action = kite.A_VALUE
				response.Sender.Id = slave

				response.Data = data

				data["name"] = response.Sender.String()
				data["description"] = iot.endpoints[response.Sender].endpoint.Description

				if err := iot.conn.WriteJSON(response); err != nil {
					iot.conn.Close()
					break
				}
				log.Printf("Response for slave --> %v", response)
			}

			//response.Data = fmt.Sprintf("New value for GPIO %s (%s) -> %t", endpoint.Description, endpoint.Address, state == 1)
			//response.Action = kite.A_LOG
			//if err := iot.conn.WriteJSON(response); err != nil {
			//	iot.conn.Close()
			//	break
			//}

		}
		iot.sync.Unlock()

	}
	log.Printf("Listening M23008 Interrupt exited (%d)...", ic.address)
}
func (ic *IC) readGPIO(gpio int) int {
	return int(mcp23008.ReadGpio(ic.ic.(*mcp23008.Mcp23008), byte(gpio)))
}

func (ic *IC) writeGPIO(gpio int, state int) int {
	if state == 0 {
		mcp23008.GpioOff(ic.ic.(*mcp23008.Mcp23008), byte(gpio))
	} else {
		mcp23008.GpioOn(ic.ic.(*mcp23008.Mcp23008), byte(gpio))
	}
	return int(mcp23008.ReadGpio(ic.ic.(*mcp23008.Mcp23008), byte(gpio)))
}
