package main

import (
	"github.com/get-code-ch/ads1115"
	kite "github.com/get-code-ch/kite-common"
	"github.com/get-code-ch/mcp23008/v3"
	"github.com/gorilla/websocket"
	"log"
	"strconv"
)

type (
	EndpointConn struct {
		conn *websocket.Conn
		endpoint kite.Endpoint
	}
)

func (iot *Iot) provisioning(data interface{}) {
	log.Printf("Provisioning iot...")

	iot.endpoints = make(map[kite.Address]*EndpointConn)
	iot.ics = make(map[string]*IC)
	for _, item := range data.([]interface{}) {
		endpoint := kite.Endpoint{}
		endpoint = endpoint.SetFromInterface(item)
		iot.endpoints[endpoint.Address] = new(EndpointConn)
		iot.endpoints[endpoint.Address].endpoint = endpoint

		// If ic is not exist create it
		if _, ok := iot.ics[endpoint.Address.Address]; !ok {
			switch endpoint.IC.Type {
			case kite.I_MCP23008:
				if icAddress, err := strconv.Atoi(endpoint.Address.Address); err == nil {
					ic := new(IC)
					ic.Address = icAddress
					ic.Type = endpoint.IC.Type
					if mcp, err := mcp23008.New(iot.conf.I2c, "", icAddress, 0, ""); err == nil {
						ic.IC = &mcp
						interrupt := make(chan byte)
						iot.ics[endpoint.Address.Address] = ic
						go mcp23008.RegisterInterrupt(ic.IC.(*mcp23008.Mcp23008),interrupt)
						go ic.listenMcp23008Interrupt(iot, interrupt)
					}
				}
				break
			case kite.I_ADS1115:
				if icAddress, err := strconv.Atoi(endpoint.Address.Address); err == nil {
					ic := new(IC)
					ic.Address = icAddress
					ic.Type = endpoint.IC.Type
					if ads, err := ads1115.New(iot.conf.I2c, "", icAddress, ""); err == nil {
						ic.IC = &ads
						iot.ics[endpoint.Address.Address] = ic
					}
				}
				break
			default:
				log.Printf("Unknown or unplemented IC")
			}
		}

		// Setting up endpoint
		switch endpoint.IC.Type {
		case kite.I_MCP23008:
			if endpoint.Attributes["mode"] == "input" || endpoint.Attributes["mode"] == "push" {
				if gpio, err := strconv.Atoi(endpoint.Address.Id); err == nil {
					if err := mcp23008.GpioSetRead(iot.ics[endpoint.Address.Address].IC.(*mcp23008.Mcp23008), byte(gpio) ); err != nil {
						log.Printf("Error configuring gpio %d as input mode --> %v", gpio, err)
					}
				}
			}
			break
		case kite.I_ADS1115:
			break
		default:
			break
		}

		// Establish connection for endpoint
		iot.endpoints[endpoint.Address].conn = connectServer(iot, endpoint.Address)

	}

}
