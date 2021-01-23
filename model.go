package main

import (
	kite "github.com/get-code-ch/kite-common"
	"sync"
)

type (
	IC struct {
		address int
		icRef   kite.IcRef
		ic      interface{}
		wg      sync.WaitGroup
		sync    sync.Mutex
	}

	SunriseSunset struct {
		Results struct {
			Sunrise                   string `json:"sunrise"`
			Sunset                    string `json:"sunset"`
			SolarNoon                 string `json:"solar_noon"`
			DayLength                 int64  `json:"day_length"`
			CivilTwilightBegin        string `json:"civil_twilight_begin"`
			CivilTwilightEnd          string `json:"civil_twilight_end"`
			NauticalTwilightBegin     string `json:"nautical_twilight_begin"`
			NauticalTwilightEnd       string `json:"nautical_twilight_end"`
			AstronomicalTwilightBegin string `json:"astronomical_twilight_begin"`
			AstronomicalTwilightEnd   string `json:"astronomical_twilight_end"`
		} `json:"results"`
		Status string `json:"status"`
	}
)
