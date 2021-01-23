package main

import (
	"github.com/get-code-ch/ads1115"
	kite "github.com/get-code-ch/kite-common"
	"log"
	"reflect"
	"time"
)

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

			data["type"] = "float"
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
