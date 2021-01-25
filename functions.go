package main

import (
	"log"
	"math"
	"reflect"
	"strings"
	"time"
)

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

// SunriseSunset function returning information about sunrise/sunset/twilight depending lat and long
func (ic *IC) SunriseSunset(inputs interface{}) string {

	// function variables
	var lat float64
	var lng float64
	var event string

	//layout := "2006-01-02T15:04:05-07:00"
	layout := "15:04:05"
	result := "N/A"

	arguments := make(map[string]interface{})

	// Check if inputs parameter are Ok, if not returning "Error" value
	if reflect.TypeOf(inputs).Kind() == reflect.TypeOf(arguments).Kind() {
		arguments = inputs.(map[string]interface{})
	} else {
		log.Printf("Invalid inputs --> %v", inputs)
		return result
	}

	// Checking inputs arguments and initializing function variables
	if input, ok := arguments["lat"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			lat = arguments["lat"].(float64)
		} else {
			return result
		}
	} else {
		return result
	}

	if input, ok := arguments["lng"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			lng = arguments["lng"].(float64)
		} else {
			return result
		}
	} else {
		return result
	}

	if input, ok := arguments["event"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.String {
			event = arguments["event"].(string)
		} else {
			return result
		}
	} else {
		return result
	}

	if sunriseSunset, err := GetSunriseSunset(lat, lng, time.Now()); err == nil {
		switch strings.ToLower(event) {
		case "sunrise":
			return sunriseSunset.Sunrise.Local().Format(layout)
		case "sunset":
			return sunriseSunset.Sunset.Local().Format(layout)
		case "twilight_begin":
			return sunriseSunset.CivilTwilightBegin.Local().Format(layout)
		case "twilight_end":
			return sunriseSunset.CivilTwilightEnd.Local().Format(layout)
		default:
			result = sunriseSunset.CivilTwilightBegin.Local().Format(layout) + "\n"
			result += sunriseSunset.Sunrise.Local().Format(layout) + "\n"
			result += sunriseSunset.Sunset.Local().Format(layout) + "\n"
			result += sunriseSunset.CivilTwilightEnd.Local().Format(layout)
			return result
		}
	} else {
		return result
	}

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
