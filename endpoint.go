package main

import (
	"fmt"
	"github.com/get-code-ch/ads1115"
	kite "github.com/get-code-ch/kite-common"
	"github.com/get-code-ch/mcp23008/v3"
	"github.com/gorilla/websocket"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

type (
	EndpointConn struct {
		conn     *websocket.Conn
		endpoint kite.Endpoint
		wg       sync.WaitGroup
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

	if data == nil {
		return
	}
	for _, item := range data.([]interface{}) {
		ic := new(IC)
		endpoint := kite.Endpoint{}
		endpoint = endpoint.SetFromInterface(item)
		iot.endpoints[endpoint.Address] = new(EndpointConn)
		iot.endpoints[endpoint.Address].endpoint = endpoint

		// If ic is not exist create it
		if _, ok := iot.ics[endpoint.Address.Address]; !ok {
			iot.ics[endpoint.Address.Address] = new(IC)
			switch endpoint.IC.Type {
			case kite.I_MCP23008:
				ic.address = endpoint.IC.Address
				ic.icRef = endpoint.IC.Type
				if mcp, err := mcp23008.New(iot.conf.I2c, endpoint.IC.Name, endpoint.IC.Address, 0, endpoint.IC.Description); err == nil {
					ic.ic = &mcp
					interrupt := make(chan byte)
					iot.ics[endpoint.Address.Address] = ic
					go mcp23008.RegisterInterrupt(ic.ic.(*mcp23008.Mcp23008), interrupt)
					go ic.listenMcp23008Interrupt(iot, interrupt)
				} else {
					log.Printf("Error creating ics --> %v", err)
				}
				break

			case kite.I_ADS1115:
				ic.address = endpoint.IC.Address
				ic.icRef = endpoint.IC.Type
				if ads, err := ads1115.New(iot.conf.I2c, endpoint.IC.Name, endpoint.IC.Address, endpoint.IC.Description); err == nil {
					ic.ic = &ads
					iot.ics[endpoint.Address.Address] = ic
				}
				break

			case kite.I_SOFT:
			case kite.I_VIRTUAL:
				break
			default:
				log.Printf("Unknown or unplemented ic")
			}
		}

		// Setting up endpoint
		switch endpoint.IC.Type {
		case kite.I_MCP23008:
			if endpoint.Attributes["mode"] == "input" || endpoint.Attributes["mode"] == "event" {
				if gpio, err := strconv.Atoi(endpoint.Address.Id); err == nil {
					if err := mcp23008.GpioSetRead(iot.ics[endpoint.Address.Address].ic.(*mcp23008.Mcp23008), byte(gpio)); err != nil {
						log.Printf("Error configuring gpio %d as input mode --> %v", gpio, err)
					}
				}
			}
			break
		case kite.I_ADS1115:
			go iot.ics[endpoint.Address.Address].refreshAds1115(iot, iot.endpoints[endpoint.Address].endpoint)
			break
		case kite.I_SOFT:
			go iot.ics[endpoint.Address.Address].refreshVirtual(iot, iot.endpoints[endpoint.Address].endpoint)
			break
		case kite.I_VIRTUAL:
			go iot.ics[endpoint.Address.Address].refreshVirtual(iot, iot.endpoints[endpoint.Address].endpoint)
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

	// When all endpoints are configured we send a log message to server and to telegram.
	iot.sync.Lock()
	response := fmt.Sprintf("%s started and provisioned", iot.conf.Name)
	message := kite.Message{Action: kite.A_LOG, Sender: iot.conf.Address, Receiver: kite.Address{Domain: iot.conf.Address.Domain, Type: kite.H_ANY, Host: "*", Address: "*", Id: "*"}, Data: response}
	_ = iot.conn.WriteJSON(message)
	message = kite.Message{Action: kite.A_NOTIFY, Sender: iot.conf.Address, Receiver: kite.Address{Domain: "telegram", Type: kite.H_ANY, Host: "*", Address: "*", Id: "*"}, Data: response}
	_ = iot.conn.WriteJSON(message)
	iot.sync.Unlock()

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
			//log.Printf("Message --> %v", message)
			cmd := ""
			gpio := 0
			state := 0
			var result interface{}

			switch message.Action {
			case kite.A_CMD:
				cmd = strings.ToLower(message.Data.(string))
				writeMode := false
				pushMode := false
				duration := 0.0

				switch ec.endpoint.Attributes["mode"].(type) {
				case string:
					writeMode = ec.endpoint.Attributes["mode"].(string) == "output"
					pushMode = ec.endpoint.Attributes["mode"].(string) == "push"
				}

				switch ec.endpoint.Attributes["duration"].(type) {
				case float64:
					duration = ec.endpoint.Attributes["duration"].(float64)
				}

				if ec.endpoint.IC.Type == kite.I_MCP23008 {
					if gpio, err = strconv.Atoi(ec.endpoint.Address.Id); err != nil {
						continue
					}

					var response = kite.Message{Sender: ec.endpoint.Address, Receiver: message.Sender, Action: kite.A_VALUE}
					switch cmd {

					// Received command is "on" force state of GPIO to 1 regardless current state
					case "on":
						if writeMode {
							state = iot.ics[ec.endpoint.Address.Address].writeGPIO(gpio, 1)
						} else {
							continue
						}
						response.Receiver = kite.Address{Domain: iot.conf.Address.Domain, Type: kite.H_ANY, Host: "*", Address: "*", Id: "*"}
						response.Action = kite.A_VALUE

						data := make(map[string]interface{})
						data["type"] = "gpio"
						data["value"] = state == 1
						data["name"] = ec.endpoint.Name
						data["description"] = ec.endpoint.Description

						response.Data = data

						_ = ec.conn.WriteJSON(response)

						break

					// Received command is "off" force state of GPIO to 0 regardless current state
					case "off":
						if writeMode {
							state = iot.ics[ec.endpoint.Address.Address].writeGPIO(gpio, 0)
						} else {
							continue
						}
						response.Receiver = kite.Address{Domain: iot.conf.Address.Domain, Type: kite.H_ANY, Host: "*", Address: "*", Id: "*"}
						response.Action = kite.A_VALUE

						data := make(map[string]interface{})
						data["type"] = "gpio"
						data["value"] = state == 1
						data["name"] = ec.endpoint.Name
						data["description"] = ec.endpoint.Description

						response.Data = data

						_ = ec.conn.WriteJSON(response)

						break

					// Received command is "reverse" we reverse  the state of GPIO
					case "reverse":
						if writeMode {
							state = int(math.Abs(float64(iot.ics[ec.endpoint.Address.Address].readGPIO(gpio) - 1)))
							state = iot.ics[ec.endpoint.Address.Address].writeGPIO(gpio, state)
						} else {
							continue
						}
						response.Receiver = kite.Address{Domain: iot.conf.Address.Domain, Type: kite.H_ANY, Host: "*", Address: "*", Id: "*"}
						response.Action = kite.A_VALUE

						data := make(map[string]interface{})
						data["type"] = "gpio"
						data["value"] = state == 1
						data["name"] = ec.endpoint.Name
						data["description"] = ec.endpoint.Description

						response.Data = data

						_ = ec.conn.WriteJSON(response)

						break

					case "push":
						if !pushMode || duration == 0 {
							continue
						}
						state = iot.ics[ec.endpoint.Address.Address].writeGPIO(gpio, 1)
						time.Sleep(time.Duration(duration) * time.Millisecond)
						state = iot.ics[ec.endpoint.Address.Address].writeGPIO(gpio, 0)

						break

					// Command "read" we reading and sending current state of GPIO
					case "read":
						state = iot.ics[ec.endpoint.Address.Address].readGPIO(gpio)
						response.Receiver = message.Sender
						response.Action = kite.A_VALUE

						data := make(map[string]interface{})
						data["type"] = "gpio"
						data["value"] = state == 1
						data["name"] = ec.endpoint.Name
						data["description"] = ec.endpoint.Description

						response.Data = data

						_ = ec.conn.WriteJSON(response)

						break
					}
					// at this point exiting switch message.Action
					continue
				}

				if ec.endpoint.IC.Type == kite.I_ADS1115 {
					var response = kite.Message{Sender: ec.endpoint.Address, Receiver: message.Sender, Action: kite.A_VALUE}
					switch cmd {

					// For ADS1115 the only one command is reading value of register
					case "read":
						result = iot.ics[ec.endpoint.Address.Address].readValueAds1115(ec.endpoint)
						response.Action = kite.A_VALUE

						data := make(map[string]interface{})

						data["type"] = "float"
						data["value"] = result
						data["name"] = ec.endpoint.Name
						data["description"] = ec.endpoint.Description

						if unit, ok := ec.endpoint.Attributes["unit"]; ok {
							data["unit"] = unit
						} else {
							data["unit"] = ""
						}

						response.Data = data

						_ = ec.conn.WriteJSON(response)

						break
					}
				}

				if ec.endpoint.IC.Type == kite.I_SOFT || ec.endpoint.IC.Type == kite.I_VIRTUAL {
					var response = kite.Message{Sender: ec.endpoint.Address, Receiver: message.Sender, Action: kite.A_VALUE}
					switch cmd {

					case "read":
						result = iot.ics[ec.endpoint.Address.Address].readValueVirtual(ec.endpoint)
						response.Action = kite.A_VALUE

						data := make(map[string]interface{})

						data["type"] = "string"
						data["value"] = result
						data["name"] = ec.endpoint.Name
						data["description"] = ec.endpoint.Description

						if unit, ok := ec.endpoint.Attributes["unit"]; ok {
							data["unit"] = unit
						} else {
							data["unit"] = ""
						}

						response.Data = data

						_ = ec.conn.WriteJSON(response)

						break
					}
				}

				break
			case kite.A_ACCEPTED:
				break
			default:
				break
			}
		}
	}
}
