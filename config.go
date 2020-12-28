package main

import (
	"encoding/json"
	"fmt"
	kite "github.com/get-code-ch/kite-common"
	"io/ioutil"
	"log"
	"os"
)

type IotConf struct {
	Name    string       `json:"name"`
	ApiKey  string       `json:"api_key"`
	Server  string       `json:"server"`
	Port    string       `json:"port"`
	Ssl     bool         `json:"ssl"`
	Address kite.Address `json:"address"`
	I2c     string       `json:"i2c"`
}

const defaultConfigFile = "./config/default.json"

func loadConfig(configFile string) *IotConf {

	// New config creation
	c := new(IotConf)

	// If no config file is provided we use "hardcoded" default filepath
	if configFile == "" {
		configFile = defaultConfigFile
	}

	// Testing if config file exist if not, return a fatal error
	_, err := os.Stat(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Panicf("Config file %s not exist\n", configFile)
		} else {
			log.Panicf("Something wrong with config file %s -> %v\n", configFile, err)
		}
	}

	// Reading and parsing configuration file
	if buffer, err := ioutil.ReadFile(configFile); err != nil {
		log.Printf(fmt.Sprintf("Error reading config file --> %v", err))
		return nil
	} else {
		if err := json.Unmarshal(buffer, c); err != nil {
			log.Panicf("Error parsing configuration file --> %v", err)
		}
		return c
	}
}
