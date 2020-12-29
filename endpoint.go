package main

import (
	"github.com/get-code-ch/ads1115"
	kite "github.com/get-code-ch/kite-common"
	"github.com/get-code-ch/mcp23008/v3"
	"github.com/gorilla/websocket"
	"log"
	"strconv"
	"strings"
	"sync"
)

type (
	EndpointConn struct {
		conn     *websocket.Conn
		endpoint kite.Endpoint
		wg       sync.WaitGroup
		//ic		 IC
	}
)

func (iot *Iot) provisioning(data interface{}) {

	//TODO: Handle reprovisioning if server is restarted
	//
	//

	var wg sync.WaitGroup
	log.Printf("Provisioning iot...")

	iot.endpoints = make(map[kite.Address]*EndpointConn)
	iot.ics = make(map[string]*IC)
	for _, item := range data.([]interface{}) {
		ic := new(IC)
		endpoint := kite.Endpoint{}
		endpoint = endpoint.SetFromInterface(item)
		iot.endpoints[endpoint.Address] = new(EndpointConn)
		iot.endpoints[endpoint.Address].endpoint = endpoint

		// If ic is not exist create it
		if _, ok := iot.ics[endpoint.Address.Address]; !ok {
			switch endpoint.IC.Type {
			case kite.I_MCP23008:
				if icAddress, err := strconv.Atoi(endpoint.Address.Address); err == nil {
					ic.address = icAddress
					ic.icRef = endpoint.IC.Type
					if mcp, err := mcp23008.New(iot.conf.I2c, "", icAddress, 0, ""); err == nil {
						ic.ic = &mcp
						interrupt := make(chan byte)
						iot.ics[endpoint.Address.Address] = ic
						go mcp23008.RegisterInterrupt(ic.ic.(*mcp23008.Mcp23008), interrupt)
						go ic.listenMcp23008Interrupt(iot, interrupt)
					}
				}
				break
			case kite.I_ADS1115:
				if icAddress, err := strconv.Atoi(endpoint.Address.Address); err == nil {
					ic.address = icAddress
					ic.icRef = endpoint.IC.Type
					if ads, err := ads1115.New(iot.conf.I2c, "", icAddress, ""); err == nil {
						ic.ic = &ads
						iot.ics[endpoint.Address.Address] = ic
					}
				}
				break
			default:
				log.Printf("Unknown or unplemented ic")
			}
		}

		// Setting up endpoint
		switch endpoint.IC.Type {
		case kite.I_MCP23008:
			if endpoint.Attributes["mode"] == "input" || endpoint.Attributes["mode"] == "push" {
				if gpio, err := strconv.Atoi(endpoint.Address.Id); err == nil {
					if err := mcp23008.GpioSetRead(iot.ics[endpoint.Address.Address].ic.(*mcp23008.Mcp23008), byte(gpio)); err != nil {
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

		wg.Add(1)
		go func() {
			for {
				// Establish connection for endpoint
				if iot.endpoints[endpoint.Address].conn = connectServer(iot, endpoint.Address); iot.endpoints[endpoint.Address].conn == nil {
					continue
				}
				iot.endpoints[endpoint.Address].wg.Add(1)
				go iot.endpoints[endpoint.Address].waitMessage(iot)
				iot.endpoints[endpoint.Address].wg.Wait()
				wg.Done()
				break
			}
		}()
	}
	wg.Wait()
}

func (ec *EndpointConn) waitMessage(iot *Iot) {
	for {
		message := kite.Message{}

		if err := ec.conn.ReadJSON(&message); err != nil {
			log.Printf("Error on readMessage -> %v", err)
			ec.wg.Done()
			return
		} else {
			switch message.Action {
			case kite.A_CMD:
				if ec.endpoint.IC.Type == kite.I_MCP23008 {
					switch strings.ToLower(message.Data.(string)) {
					case "on":
						if gpio, err := strconv.Atoi(ec.endpoint.Address.Id); err == nil {
							iot.ics[ec.endpoint.Address.Address].writeGPIO(gpio, 1)
						}
						break
					case "off":
						if strings.ToLower(message.Data.(string)) == "off" {
							if gpio, err := strconv.Atoi(ec.endpoint.Address.Id); err == nil {
								iot.ics[ec.endpoint.Address.Address].writeGPIO(gpio, 0)
							}
							break
						}
					}
				}
				break
			case kite.A_ACCEPTED:
				break
			default:
				//log.Printf("Message received -> %v", message.Data)
				break
			}
		}
	}
}
