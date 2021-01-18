package main

import (
	"fmt"
	"github.com/get-code-ch/ads1115"
	kite "github.com/get-code-ch/kite-common"
	"github.com/get-code-ch/mcp23008/v3"
	"log"
	"math"
	"reflect"
	"strconv"
	"sync"
	"time"
)

type (
	IC struct {
		address int
		icRef   kite.IcRef
		ic      interface{}
		wg      sync.WaitGroup
		sync    sync.Mutex
	}
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

func (ic *IC) refreshAds1115(iot *Iot, endpoint kite.Endpoint) {
	var refreshRate time.Duration
	if val, ok := endpoint.Attributes["refresh_interval"]; ok {
		refreshRate = time.Duration(val.(float64)) * time.Second
	} else {
		return
	}

	ticker := time.NewTicker(refreshRate).C

	var response = kite.Message{
		Sender:   endpoint.Address,
		Receiver: kite.Address{Domain: endpoint.Address.Domain, Type: "*", Host: "*", Address: "*", Id: "*"},
		Action:   kite.A_VALUE,
		Data:     nil,
	}

	for {
		select {
		case <-ticker:
			log.Printf("Refresh event for %s", endpoint.Description)
			result := ic.readValueAds1115(endpoint)
			data := make(map[string]interface{})

			data["type"] = "analog"
			data["value"] = result
			data["unit"] = endpoint.Attributes["unit"]
			data["name"] = endpoint.Name
			data["description"] = endpoint.Description

			response.Data = data

			iot.sync.Lock()
			if err := iot.conn.WriteJSON(response); err != nil {
				iot.conn.Close()
				iot.sync.Unlock()
				break
			}
			iot.sync.Unlock()

		}
	}
}

func (ic *IC) readValueAds1115(endpoint kite.Endpoint) interface{} {
	ads := ic.ic.(*ads1115.Ads1115)
	ic.sync.Lock()
	vIn := ads1115.ReadConversionRegister(ads, endpoint.Address.Id)
	ic.sync.Unlock()
	result := 0.0

	if _, ok := endpoint.Attributes["scale"]; ok {
		if _, ok := endpoint.Attributes["convert"]; ok {
			fnc := reflect.ValueOf(ic).MethodByName(endpoint.Attributes["convert"].(string))
			if fnc.IsValid() {
				arguments := endpoint.Attributes
				arguments["vIn"] = vIn
				inputs := make([]reflect.Value, 1)

				inputs[0] = reflect.ValueOf(arguments)
				result = fnc.Call(inputs)[0].Float()
			} else {
				log.Printf("Converting function %s doesn't exist", endpoint.Attributes["convert"].(string))
			}
		} else {
			result = vIn * endpoint.Attributes["scale"].(float64)
		}
	}
	return result
}

// OhmMeter function returning calculated value of resistance
func (ic *IC) OhmMeter(inputs interface{}) float64 {

	// function variables
	var vIn float64
	var vcc float64
	var result float64
	var scale float64

	arguments := make(map[string]interface{})
	result = -1.0
	scale = 1

	// Check if inputs parameter are Ok, if not returning "Error" value
	if reflect.TypeOf(inputs).Kind() == reflect.TypeOf(arguments).Kind() {
		arguments = inputs.(map[string]interface{})
	} else {
		log.Printf("Invalid inputs --> %v", inputs)
		return result
	}

	// Checking inputs arguments and initializing function variables
	if input, ok := arguments["vIn"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			vIn = arguments["vIn"].(float64)
		} else {
			return result
		}
	} else {
		return result
	}

	if input, ok := arguments["scale"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			scale = arguments["scale"].(float64)
		}
	}

	if input, ok := arguments["vcc"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			vcc = arguments["vcc"].(float64)
		} else {
			return result
		}
	} else {
		return result
	}

	// Calculating Ohm value
	if reference, ok := arguments["reference"]; ok {
		if reflect.TypeOf(reference).Kind() == reflect.Float64 {
			result = ((vcc/vIn - 1) * reference.(float64)) * scale
		}
	}
	return result

}

func (ic *IC) ToLux(inputs interface{}) float64 {

	// function variables
	var vIn float64
	var result float64
	var scale float64

	arguments := make(map[string]interface{})
	scale = 1
	result = -1.0

	// Check if inputs parameter are Ok, if not returning "Error" value
	if reflect.TypeOf(inputs).Kind() == reflect.TypeOf(arguments).Kind() {
		arguments = inputs.(map[string]interface{})
	} else {
		log.Printf("Invalid inputs --> %v", inputs)
		return result
	}

	// Checking inputs arguments and initializing function variables
	if input, ok := arguments["scale"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			scale = arguments["scale"].(float64)
		}
	}

	if input, ok := arguments["vIn"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			vIn = arguments["vIn"].(float64)
		} else {
			return result
		}
	} else {
		return result
	}

	// Calculating resulting value
	result = (vIn * (700 + math.Log10(vIn)*100)) * scale
	if math.IsNaN(result) {
		result = -1
	}
	return result

}

func (ic *IC) readValueVirtual(endpoint kite.Endpoint) interface{} {

	var result interface{}
	if _, ok := endpoint.Attributes["function"]; ok {
		fnc := reflect.ValueOf(ic).MethodByName(endpoint.Attributes["function"].(string))
		if fnc.IsValid() {
			arguments := endpoint.Attributes
			inputs := make([]reflect.Value, 1)

			inputs[0] = reflect.ValueOf(arguments)
			result = fnc.Call(inputs)[0]
		} else {
			log.Printf("Converting function %s doesn't exist", endpoint.Attributes["function"].(string))
		}
	} else {
		result = nil
	}
	return result
}

// SunriseSunset function returning information about sunrise/sunset/twilight depending lat and long
func (ic *IC) SunriseSunset(inputs interface{}) interface{} {

	// function variables
	var lat float64
	var lng float64

	arguments := make(map[string]interface{})

	// Check if inputs parameter are Ok, if not returning "Error" value
	if reflect.TypeOf(inputs).Kind() == reflect.TypeOf(arguments).Kind() {
		arguments = inputs.(map[string]interface{})
	} else {
		log.Printf("Invalid inputs --> %v", inputs)
		return nil
	}

	// Checking inputs arguments and initializing function variables
	if input, ok := arguments["lat"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			lat = arguments["lat"].(float64)
		} else {
			return nil
		}
	} else {
		return nil
	}

	if input, ok := arguments["lng"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			lng = arguments["lng"].(float64)
		} else {
			return nil
		}
	} else {
		return nil
	}

	// Getting sunrise and sunset information
	_ = lat
	_ = lng

	return nil

}

//---------------------------------------------------------------------------
// template function returning calculated value of
func (ic *IC) template(inputs interface{}) float64 {

	// function variables
	var vIn float64
	var result float64
	var scale float64

	arguments := make(map[string]interface{})
	scale = 1
	result = -1.0

	// Check if inputs parameter are Ok, if not returning "Error" value
	if reflect.TypeOf(inputs).Kind() == reflect.TypeOf(arguments).Kind() {
		arguments = inputs.(map[string]interface{})
	} else {
		return result
	}

	// Checking inputs arguments and initializing function variables
	if input, ok := arguments["scale"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			scale = arguments["scale"].(float64)
		}
	}

	if input, ok := arguments["vIn"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			vIn = arguments["vIn"].(float64)
		} else {
			return result
		}
	} else {
		return result
	}

	/*
		if input, ok := arguments["vcc"]; ok {
			if reflect.TypeOf(input).Kind() == reflect.Float64 {
				vcc = arguments["vcc"].(float64)
			} else {
				return result
			}
		} else {
			return result
		}
	*/
	// Calculating resulting value
	result = vIn * scale
	return result

}
