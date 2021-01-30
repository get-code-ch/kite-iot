package main

import (
	kite "github.com/get-code-ch/kite-common"
	"log"
	"reflect"
	"time"
)

func (ic *IC) readValueVirtual(endpoint kite.Endpoint) interface{} {

	result := ""
	if _, ok := endpoint.Attributes["function"]; ok {
		fnc := reflect.ValueOf(ic).MethodByName(endpoint.Attributes["function"].(string))
		if fnc.IsValid() {
			arguments := endpoint.Attributes
			inputs := make([]reflect.Value, 1)

			inputs[0] = reflect.ValueOf(arguments)
			result = fnc.Call(inputs)[0].String()
		} else {
			log.Printf("Converting function %s doesn't exist", endpoint.Attributes["function"].(string))
		}
	}
	return result
}

func (ic *IC) refreshVirtual(iot *Iot, endpoint kite.Endpoint) {
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
			result := ic.readValueVirtual(endpoint)
			data := make(map[string]interface{})

			data["type"] = "string"
			data["value"] = result
			data["name"] = endpoint.Name
			data["description"] = endpoint.Description

			if unit, ok := endpoint.Attributes["unit"]; ok {
				data["unit"] = unit
			} else {
				data["unit"] = ""
			}

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
