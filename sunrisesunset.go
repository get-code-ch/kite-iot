package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// SunriseSunset function returning information about sunrise/sunset/twilight depending lat and long
// this function return time of sunrise/sunset/twilight start/twilight end
// to limit access to sunrise-sunset.org API requests are cached in temporary files if corresponding request is already made
// we read temporary files otherwise we execute get api
func GetSunriseSunset(lat float64, lng float64, date time.Time) (SunriseSunset, error) {

	var sunriseSunset SunriseSunset
	var sunriseSunsetRaw SunriseSunsetRaw
	var data []byte

	layout := "2006-01-02T15:04:05-07:00"

	// Getting sunrise and sunset information by request to sunrise-sunset.org
	// https://api.sunrise-sunset.org/json?lat=36.7201600&lng=-4.4203400&date=today

	// Formatting name of caches files
	current := date.Format("2006-01-02")
	fileCurrent := fmt.Sprintf("%s/kite-iot_%s_%.4f_%.4f.json", os.TempDir(), current, lat, lng)

	if _, err := os.Stat(fileCurrent); err != nil {
		client := new(http.Client)
		if file, err := os.Create(fileCurrent); err == nil {
			if response, err := client.Get(fmt.Sprintf("https://api.sunrise-sunset.org/json?lat=%.4f&lng=%.4f&date=%s&formatted=0", lat, lng, current)); err == nil {
				if body, err := ioutil.ReadAll(response.Body); err == nil {
					file.Write(body)
					file.Close()
					data = make([]byte, len(body))
					copy(data, body)
				}
			}
		}
	} else {
		if content, err := ioutil.ReadFile(fileCurrent); err == nil {
			data = make([]byte, len(content))
			copy(data, content)
		}
	}

	if err := json.Unmarshal(data, &sunriseSunsetRaw); err == nil {
		if value, err := time.Parse(layout, sunriseSunsetRaw.Results.Sunrise); err == nil {
			sunriseSunset.Sunrise = value
		}
		if value, err := time.Parse(layout, sunriseSunsetRaw.Results.Sunset); err == nil {
			sunriseSunset.Sunset = value
		}
		if value, err := time.Parse(layout, sunriseSunsetRaw.Results.SolarNoon); err == nil {
			sunriseSunset.SolarNoon = value
		}
		sunriseSunset.DayLength = sunriseSunsetRaw.Results.DayLength
		if value, err := time.Parse(layout, sunriseSunsetRaw.Results.CivilTwilightBegin); err == nil {
			sunriseSunset.CivilTwilightBegin = value
		}
		if value, err := time.Parse(layout, sunriseSunsetRaw.Results.CivilTwilightEnd); err == nil {
			sunriseSunset.CivilTwilightEnd = value
		}
		if value, err := time.Parse(layout, sunriseSunsetRaw.Results.NauticalTwilightBegin); err == nil {
			sunriseSunset.NauticalTwilightBegin = value
		}
		if value, err := time.Parse(layout, sunriseSunsetRaw.Results.NauticalTwilightEnd); err == nil {
			sunriseSunset.NauticalTwilightEnd = value
		}
		if value, err := time.Parse(layout, sunriseSunsetRaw.Results.AstronomicalTwilightBegin); err == nil {
			sunriseSunset.AstronomicalTwilightBegin = value
		}
		if value, err := time.Parse(layout, sunriseSunsetRaw.Results.AstronomicalTwilightEnd); err == nil {
			sunriseSunset.AstronomicalTwilightEnd = value
		}
		return sunriseSunset, nil
	} else {
		return SunriseSunset{}, err
	}
}
