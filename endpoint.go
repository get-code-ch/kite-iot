package main

import (
	"github.com/get-code-ch/ads1115"
	kite "github.com/get-code-ch/kite-common"
	"github.com/get-code-ch/mcp23008/v3"
	"log"
	"strconv"
)

func (iot *Iot) provisioning(data interface{}) {
	log.Printf("Provisioning iot...")

	iot.endpoints = nil
	iot.ics = make(map[string]*IC)
	for _, item := range data.([]interface{}) {
		endpoint := kite.Endpoint{}
		endpoint = endpoint.SetFromInterface(item)
		iot.endpoints = append(iot.endpoints, endpoint)

		// If ic not already configured create it
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
					mcp23008.GpioSetRead(iot.ics[endpoint.Address.Address].IC.(*mcp23008.Mcp23008), byte(gpio) )
				}
			}
			break
		case kite.I_ADS1115:
			break
		default:
			break
		}
	}

}
