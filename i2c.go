package main

import (
	"fmt"
	"github.com/get-code-ch/ads1115"
	kite "github.com/get-code-ch/kite-common"
	"github.com/get-code-ch/mcp23008/v3"
	"log"
	"math"
	"reflect"
	"sync"
)

type (
	IC struct {
		address int
		icRef   kite.IcRef
		ic      interface{}
		wg      sync.WaitGroup
		sync	sync.Mutex
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

		iot.sync.Lock()
		var message = kite.Message{Data: fmt.Sprintf("%s (%s) is updated to %d", endpoint.Description, endpoint.Address, state), Sender: address, Receiver: kite.Address{Domain: iot.conf.Address.Domain, Type: kite.H_ANY, Host: "*", Address: "*", Id: "*"}, Action: kite.A_NOTIFY}
		if err := iot.conn.WriteJSON(message); err != nil {
			iot.conn.Close()
			break
		}

		message.Action = kite.A_LOG
		if err := iot.conn.WriteJSON(message); err != nil {
			iot.conn.Close()
			break
		}

		if endpoint.Notification.Telegram {
			message.Action = kite.A_NOTIFY
			message.Receiver = kite.Address{Domain: "telegram", Type: kite.H_ANY, Host: "*", Address: "*", Id: "*"}
			if err := iot.conn.WriteJSON(message); err != nil {
				iot.conn.Close()
				break
			}
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

func (ic *IC) readValue(endpoint kite.Endpoint) float64 {
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

	return result

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
