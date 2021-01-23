package main

import (
	kite "github.com/get-code-ch/kite-common"
	"log"
	"reflect"
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
