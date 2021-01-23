package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
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

	var sunriseSunset SunriseSunset

	arguments := make(map[string]interface{})

	// Check if inputs parameter are Ok, if not returning "Error" value
	if reflect.TypeOf(inputs).Kind() == reflect.TypeOf(arguments).Kind() {
		arguments = inputs.(map[string]interface{})
	} else {
		log.Printf("Invalid inputs --> %v", inputs)
		return "N/A"
	}

	// Checking inputs arguments and initializing function variables
	if input, ok := arguments["lat"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			lat = arguments["lat"].(float64)
		} else {
			return "N/A"
		}
	} else {
		return "N/A"
	}

	if input, ok := arguments["lng"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			lng = arguments["lng"].(float64)
		} else {
			return "N/A"
		}
	} else {
		return "N/A"
	}

	if input, ok := arguments["event"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.String {
			event = arguments["event"].(string)
		} else {
			return "N/A"
		}
	} else {
		return "N/A"
	}

	// Getting sunrise and sunset information by request to sunrise-sunset.org
	// https://api.sunrise-sunset.org/json?lat=36.7201600&lng=-4.4203400&date=today

	//TODO: During devs we don't make request to API provider
	body := `{"results":{"sunrise":"2021-01-22T07:07:05+00:00","sunset":"2021-01-22T16:23:02+00:00","solar_noon":"2021-01-22T11:45:03+00:00","day_length":33357,"civil_twilight_begin":"2021-01-22T06:34:00+00:00","civil_twilight_end":"2021-01-22T16:56:06+00:00","nautical_twilight_begin":"2021-01-22T05:57:04+00:00","nautical_twilight_end":"2021-01-22T17:33:03+00:00","astronomical_twilight_begin":"2021-01-22T05:21:15+00:00","astronomical_twilight_end":"2021-01-22T18:08:51+00:00"},"status":"OK"}`
	if err := json.Unmarshal([]byte(body), &sunriseSunset); err == nil {
		//+++
		switch strings.ToLower(event) {
		case "sunrise":
			layout := "2006-01-02T15:04:05-07:00"
			if value, err := time.Parse(layout, sunriseSunset.Results.Sunrise); err == nil {
				return value.UTC().Local().Format("15:04:05")
			} else {
				return sunriseSunset.Results.Sunrise
			}
		case "sunset":
			layout := "2006-01-02T15:04:05-07:00"
			if value, err := time.Parse(layout, sunriseSunset.Results.Sunset); err == nil {
				return value.UTC().Local().Format("15:04:05")
			} else {
				return sunriseSunset.Results.Sunset
			}
		case "twilight_begin":
			layout := "2006-01-02T15:04:05-07:00"
			if value, err := time.Parse(layout, sunriseSunset.Results.CivilTwilightBegin); err == nil {
				return value.UTC().Local().Format("15:04:05")
			} else {
				return sunriseSunset.Results.CivilTwilightBegin
			}
		case "twilight_end":
			layout := "2006-01-02T15:04:05-07:00"
			if value, err := time.Parse(layout, sunriseSunset.Results.CivilTwilightEnd); err == nil {
				return value.UTC().Local().Format("15:04:05")
			} else {
				return sunriseSunset.Results.CivilTwilightEnd
			}
		default:
			return "N/A"
		}
		//---
		//return sunriseSunset
	} else {
		return "N/A"
	}

	//TODO: During devs this part of code is never accessed
	client := new(http.Client)
	if response, err := client.Get(fmt.Sprintf("https://api.sunrise-sunset.org/json?lat=%.4f&lng=%.4f&date=%s&formatted=0", lat, lng, time.Now().Format("2006-01-02"))); err == nil {
		if body, err := ioutil.ReadAll(response.Body); err == nil {
			if err := json.Unmarshal(body, &sunriseSunset); err == nil {
				//+++
				switch strings.ToLower(event) {
				case "sunrise":
					return sunriseSunset.Results.Sunrise
				default:
					return "N/A"
				}
				//---
			}
		}
	} else {
		log.Printf("Error getting sunrise, sunset --> %v", err)
	}

	return "N/A"

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
